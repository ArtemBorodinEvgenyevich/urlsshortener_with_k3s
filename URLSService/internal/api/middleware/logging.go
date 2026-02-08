package middleware

import (
	"net/http"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Logging() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())

			requestLog := logger.L().With(
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("proto", r.Proto),
				zap.String("host", r.Host),
			)

			ctx := logger.WithContext(r.Context(), requestLog)
			requestLog.Info("Request Started")

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))

			requestLog.Info("Request completed",
				zap.Int("status", ww.Status()),
				zap.String("status_text", http.StatusText(ww.Status())),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
