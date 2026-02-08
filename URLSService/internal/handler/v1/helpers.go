package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"go.uber.org/zap"
)

// respondWithJSON is a helper function for consistent JSON responses
func respondWithJSON(ctx context.Context, w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.AppLogErrorCtx(ctx, "Failed to encode JSON response", zap.Error(err))
		}
	}
}

// respondWithError is a helper function for consistent error responses
func respondWithError(ctx context.Context, w http.ResponseWriter, statusCode int, error string, message string) {
	errorResponse := ErrorResponse{
		Error:   error,
		Message: message,
	}
	respondWithJSON(ctx, w, statusCode, errorResponse)
}
