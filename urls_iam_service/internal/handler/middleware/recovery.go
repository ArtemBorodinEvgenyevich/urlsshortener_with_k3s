package middleware

import (
	"net/http"
	"runtime/debug"
	"urls_iam_service/internal/pkg/response"

	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log panic
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.ByteString("stack", debug.Stack()),
					)

					// Return 500
					response.Error(w, logger, http.StatusInternalServerError, nil, "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
