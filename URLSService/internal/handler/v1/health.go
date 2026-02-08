package v1

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthHandler struct {
	pgPool       *pgxpool.Pool
	redisClient  *redis.Client
	shuttingDown *atomic.Bool
}

func NewHealthHandler(pgPool *pgxpool.Pool, redisClient *redis.Client, shuttingDown *atomic.Bool) *HealthHandler {
	return &HealthHandler{
		pgPool:       pgPool,
		redisClient:  redisClient,
		shuttingDown: shuttingDown,
	}
}

// HealthCheck returns the basic health status of the service
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}
	respondWithJSON(r.Context(), w, http.StatusOK, response)
}

// ReadinessCheck checks if the service is ready to accept traffic
// It verifies database connections and checks if the service is shutting down
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if service is shutting down
	if h.shuttingDown.Load() {
		response := ReadinessResponse{
			Status: "unavailable",
			Reason: "shutting down",
		}
		respondWithJSON(ctx, w, http.StatusServiceUnavailable, response)
		return
	}

	opCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := ReadinessResponse{
		Status: "ok",
	}

	// Check PostgreSQL connection
	if err := h.pgPool.Ping(opCtx); err != nil {
		logger.PgLogWarnCtx(ctx, "PostgreSQL health check failed", zap.Error(err))
		response.Status = "unavailable"
		response.Postgres = "down"
		respondWithJSON(ctx, w, http.StatusServiceUnavailable, response)
		return
	}
	response.Postgres = "up"

	// Check Redis connection
	if err := h.redisClient.Ping(opCtx).Err(); err != nil {
		logger.RedisLogWarnCtx(ctx, "Redis health check failed", zap.Error(err))
		response.Status = "unavailable"
		response.Redis = "down"
		respondWithJSON(ctx, w, http.StatusServiceUnavailable, response)
		return
	}
	response.Redis = "up"

	respondWithJSON(ctx, w, http.StatusOK, response)
}
