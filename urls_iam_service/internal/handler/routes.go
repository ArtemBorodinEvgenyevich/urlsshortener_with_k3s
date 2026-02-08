package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	customMiddleware "urls_iam_service/internal/handler/middleware"
)

type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Health *HealthHandler
}

// SetupRouter sets up HTTP routes
func SetupRouter(handlers *Handlers, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(customMiddleware.Recovery(logger))
	r.Use(customMiddleware.Logging(logger))
	r.Use(customMiddleware.CORS([]string{"*"})) // Configure properly in production
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health endpoints
	r.Get("/health", handlers.Health.Health)
	r.Head("/health", handlers.Health.Health)
	r.Get("/ready", handlers.Health.Ready)
	r.Head("/ready", handlers.Health.Ready)

	// Auth endpoints
	r.Route("/auth", func(r chi.Router) {
		r.Get("/validate", handlers.Auth.Validate)              // ForwardAuth endpoint
		r.Post("/session", handlers.Auth.CreateSession)         // Create/validate session
		r.Put("/session/refresh", handlers.Auth.RefreshSession) // Refresh session
		r.Delete("/logout", handlers.Auth.Logout)               // Logout
	})

	// User endpoints
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", handlers.User.GetUser) // Get user by ID
	})

	return r
}
