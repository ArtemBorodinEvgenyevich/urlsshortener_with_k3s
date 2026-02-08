package handler

import (
	"errors"
	"net/http"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/handler/dto"
	"urls_iam_service/internal/pkg/response"
	"urls_iam_service/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userService service.UserService,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from URL
	userID := chi.URLParam(r, "id")
	if userID == "" {
		response.Error(w, h.logger, http.StatusBadRequest, nil, "User ID is required")
		return
	}

	// Get user
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, h.logger, http.StatusNotFound, err, "User not found")
			return
		}
		response.Error(w, h.logger, http.StatusInternalServerError, err, "Failed to get user")
		return
	}

	// Return response
	resp := dto.UserResponse{
		ID:         user.ID,
		Provider:   string(user.Provider),
		ProviderID: user.ProviderID,
		Metadata:   user.Metadata,
		CreatedAt:  user.CreatedAt,
		LastSeenAt: user.LastSeenAt,
	}

	response.JSON(w, http.StatusOK, resp)
}
