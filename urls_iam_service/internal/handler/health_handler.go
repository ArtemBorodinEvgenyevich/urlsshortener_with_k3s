package handler

import (
	"context"
	"net/http"
	"time"
	"urls_iam_service/internal/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	db *pgxpool.Pool,
	redisClient *redis.Client,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services,omitempty"`
}

// Health handles GET /health (liveness probe)
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status: "ok",
	}
	response.JSON(w, http.StatusOK, resp)
}

// Ready handles GET /ready (readiness probe)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("postgres health check failed", zap.Error(err))
		services["postgres"] = "unhealthy"
		allHealthy = false
	} else {
		services["postgres"] = "healthy"
	}

	// Check Redis
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		h.logger.Error("redis health check failed", zap.Error(err))
		services["redis"] = "unhealthy"
		allHealthy = false
	} else {
		services["redis"] = "healthy"
	}

	resp := HealthResponse{
		Status:   "ok",
		Services: services,
	}

	statusCode := http.StatusOK
	if !allHealthy {
		resp.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	response.JSON(w, statusCode, resp)
}
