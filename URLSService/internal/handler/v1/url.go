package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

// Create creates a new short URL
func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.AppLogErrorCtx(ctx, "Failed to decode request body", zap.Error(err))
		respondWithError(ctx, w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from X-User-Id header (set by Traefik ForwardAuth)
	var userID *string
	if uid := r.Header.Get("X-User-Id"); uid != "" {
		userID = &uid
	}

	url, err := h.service.CreateShortURL(ctx, req.URL, req.TTL, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			logger.AppLogInfoCtx(ctx, "Invalid URL provided",
				zap.String("url", req.URL),
				zap.Error(err),
			)
			respondWithError(ctx, w, http.StatusBadRequest, "Invalid URL", err.Error())
		case errors.Is(err, service.ErrInvalidTTL):
			logger.AppLogInfoCtx(ctx, "Invalid URL provided",
				zap.Int("ttl", req.TTL),
				zap.Error(err),
			)
			respondWithError(ctx, w, http.StatusBadRequest, "Invalid TTL", err.Error())
		default:
			logger.AppLogErrorCtx(ctx, "Failed to create URL",
				zap.Error(err),
				zap.String("url", req.URL),
			)
			respondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", "")
		}
		return
	}

	response := URLResponse{
		ShortURL:  url.ShortCode,
		ExpiresAt: url.ExpiresAt.Format(time.RFC3339),
	}

	respondWithJSON(ctx, w, http.StatusCreated, response)
}

// Get retrieves the original URL by short code
func (h *URLHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		logger.AppLogInfoCtx(ctx, "Empty short code provided")
		respondWithError(ctx, w, http.StatusBadRequest, "Short code is required", "")
		return
	}

	url, err := h.service.GetURL(ctx, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			logger.AppLogInfoCtx(ctx, "URL not found",
				zap.String("short_code", shortCode),
			)
			respondWithError(ctx, w, http.StatusNotFound, "URL not found", "")
		default:
			logger.AppLogErrorCtx(ctx, "Failed to get URL",
				zap.Error(err),
				zap.String("short_code", shortCode),
			)
			respondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", "")
		}
		return
	}

	response := URLDataResponse{
		OriginalURL: url.OriginalURL,
		ExpiresAt:   url.ExpiresAt.Format(time.RFC3339),
	}

	respondWithJSON(ctx, w, http.StatusOK, response)
}

// List returns all URLs for the current user
func (h *URLHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from X-User-Id header (set by Traefik ForwardAuth)
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		logger.AppLogInfoCtx(ctx, "No user ID provided for list operation")
		respondWithError(ctx, w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Parse query params
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	urls, err := h.service.GetUserURLs(ctx, userID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			respondWithError(ctx, w, http.StatusForbidden, "Access denied", "")
		default:
			logger.AppLogErrorCtx(ctx, "Failed to get user URLs",
				zap.Error(err),
				zap.String("user_id", userID),
			)
			respondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", "")
		}
		return
	}

	items := make([]URLListItem, 0, len(urls))
	for _, url := range urls {
		items = append(items, URLListItem{
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
			ExpiresAt:   url.ExpiresAt.Format(time.RFC3339),
			CreatedAt:   url.CreatedAt.Format(time.RFC3339),
		})
	}

	response := URLListResponse{URLs: items}
	respondWithJSON(ctx, w, http.StatusOK, response)
}

// Delete removes a short URL
func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		logger.AppLogInfoCtx(ctx, "Empty short code provided")
		respondWithError(ctx, w, http.StatusBadRequest, "Short code is required", "")
		return
	}

	// Get user ID from X-User-Id header (set by Traefik ForwardAuth)
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		logger.AppLogInfoCtx(ctx, "No user ID provided for delete operation")
		respondWithError(ctx, w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	err := h.service.DeleteURL(ctx, shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			logger.AppLogInfoCtx(ctx, "URL not found",
				zap.String("short_code", shortCode),
			)
			respondWithError(ctx, w, http.StatusNotFound, "URL not found", "")
		case errors.Is(err, service.ErrForbidden):
			logger.AppLogInfoCtx(ctx, "Access denied for URL deletion",
				zap.String("short_code", shortCode),
				zap.String("user_id", userID),
			)
			respondWithError(ctx, w, http.StatusForbidden, "Access denied", "")
		default:
			logger.AppLogErrorCtx(ctx, "Failed to delete URL",
				zap.Error(err),
				zap.String("short_code", shortCode),
			)
			respondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", "")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
