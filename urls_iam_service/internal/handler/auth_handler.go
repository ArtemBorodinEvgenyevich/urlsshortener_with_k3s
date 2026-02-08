package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"urls_iam_service/internal/config"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/handler/dto"
	"urls_iam_service/internal/pkg/response"
	"urls_iam_service/internal/service"

	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
	cookieCfg   config.CookieConfig
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authService service.AuthService,
	cookieCfg config.CookieConfig,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cookieCfg:   cookieCfg,
		logger:      logger,
	}
}

// Validate handles GET /auth/validate (ForwardAuth endpoint)
func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	// Extract session_id from cookie
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		h.logger.Debug("no session_id in cookie")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Validate session
	sessionInfo, err := h.authService.ValidateSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrSessionExpired) {
			h.logger.Debug("session invalid",
				zap.String("session_id", sessionID),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.logger.Error("failed to validate session",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers for Traefik
	w.Header().Set("X-User-Id", sessionInfo.UserID)
	w.Header().Set("X-Provider", string(sessionInfo.Provider))
	w.WriteHeader(http.StatusOK)
}

// CreateSession handles POST /auth/session
func (h *AuthHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req dto.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, h.logger, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Extract existing session_id from cookie (if any)
	existingSessionID := h.extractSessionID(r)

	// Create or validate session
	result, err := h.authService.CreateOrValidateSession(r.Context(), existingSessionID, req.Metadata)
	if err != nil {
		response.Error(w, h.logger, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, result.SessionID, result.ExpiresAt)

	// Return response
	resp := dto.CreateSessionResponse{
		UserID:    result.UserID,
		SessionID: result.SessionID,
		IsNewUser: result.IsNewUser,
		ExpiresAt: result.ExpiresAt,
		Provider:  string(result.Provider),
	}

	statusCode := http.StatusOK
	if result.IsNewUser {
		statusCode = http.StatusCreated
	}

	response.JSON(w, statusCode, resp)
}

// RefreshSession handles PUT /auth/session/refresh
func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	// Extract session_id from cookie
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		response.Error(w, h.logger, http.StatusUnauthorized, nil, "No session found")
		return
	}

	// Refresh session
	expiresAt, err := h.authService.RefreshSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			response.Error(w, h.logger, http.StatusNotFound, err, "Session not found")
			return
		}
		if errors.Is(err, domain.ErrSessionExpired) {
			response.Error(w, h.logger, http.StatusUnauthorized, err, "Session expired")
			return
		}
		response.Error(w, h.logger, http.StatusInternalServerError, err, "Failed to refresh session")
		return
	}

	// Update cookie with new expiration
	h.setSessionCookie(w, sessionID, expiresAt)

	// Return response
	resp := dto.RefreshSessionResponse{
		ExpiresAt: expiresAt,
	}

	response.JSON(w, http.StatusOK, resp)
}

// Logout handles DELETE /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract session_id from cookie
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		response.NoContent(w)
		return
	}

	// Logout
	if err := h.authService.Logout(r.Context(), sessionID); err != nil {
		h.logger.Error("failed to logout", zap.String("session_id", sessionID), zap.Error(err))
		// Don't return error to user, just delete cookie
	}

	// Delete cookie
	h.deleteSessionCookie(w)

	response.NoContent(w)
}

// Helper: extract session_id from cookie
func (h *AuthHandler) extractSessionID(r *http.Request) string {
	cookie, err := r.Cookie(h.cookieCfg.Name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// Helper: set session cookie
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     h.cookieCfg.Name,
		Value:    sessionID,
		Path:     h.cookieCfg.Path,
		Domain:   h.cookieCfg.Domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   h.cookieCfg.Secure,
		HttpOnly: h.cookieCfg.HttpOnly,
		SameSite: h.parseSameSite(h.cookieCfg.SameSite),
	}

	http.SetCookie(w, cookie)
}

// Helper: delete session cookie
func (h *AuthHandler) deleteSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     h.cookieCfg.Name,
		Value:    "",
		Path:     h.cookieCfg.Path,
		Domain:   h.cookieCfg.Domain,
		MaxAge:   -1,
		Secure:   h.cookieCfg.Secure,
		HttpOnly: h.cookieCfg.HttpOnly,
		SameSite: h.parseSameSite(h.cookieCfg.SameSite),
	}

	http.SetCookie(w, cookie)
}

// Helper: parse SameSite string to http.SameSite
func (h *AuthHandler) parseSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
