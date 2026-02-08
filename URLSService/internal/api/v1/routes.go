package v1

import (
	"sync/atomic"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/api/middleware"
	v1 "github.com/ArtemBorodinEvgenyevich/URLSService/internal/handler/v1"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config holds dependencies needed for v1 API routes
type Config struct {
	URLService   service.URLService
	PgPool       *pgxpool.Pool
	RedisClient  *redis.Client
	ShuttingDown *atomic.Bool
}

// RegisterRoutes registers all v1 API routes
func RegisterRoutes(r chi.Router, cfg *Config) {
	// Initialize handlers
	urlHandler := v1.NewURLHandler(cfg.URLService)
	healthHandler := v1.NewHealthHandler(cfg.PgPool, cfg.RedisClient, cfg.ShuttingDown)

	// API v1 group
	r.Route("/api/v1", func(r chi.Router) {
		// Apply v1-specific middleware
		r.Use(middleware.APIVersion("v1"))
		r.Use(middleware.CORS())

		// Health endpoints (not versioned in path, but under /api/v1)
		r.Get("/health", healthHandler.HealthCheck)
		r.Get("/readiness", healthHandler.ReadinessCheck)

		// URL shortener endpoints
		r.Post("/shorten", urlHandler.Create)
		r.Get("/urls", urlHandler.List)
		r.Get("/urls/{shortCode}", urlHandler.Get)
		r.Delete("/urls/{shortCode}", urlHandler.Delete)
	})
}
