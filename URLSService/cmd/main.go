package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	api_middleware "github.com/ArtemBorodinEvgenyevich/URLSService/internal/api/middleware"
	apiv1 "github.com/ArtemBorodinEvgenyevich/URLSService/internal/api/v1"
	cache_redis "github.com/ArtemBorodinEvgenyevich/URLSService/internal/cache/redis"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/config"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/repository"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/repository/postgres"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	shutdownPeriod      = 15 * time.Second
	shutdownHardPeriod  = 3 * time.Second
	readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	// Logger initialize
	if err := logger.Initialize(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.AppLogError("Logger sync error, can't flush buffered logs")
		}
	}()

	logger.AppLogInfo("URL Shortener starting")

	// Load environment variables
	envFilePath := filepath.Join("docker", ".env")
	if err := godotenv.Load(envFilePath); err != nil {
		logger.AppLogWarn("Warning: .env file not found, using system environment variables")
	}

	// Load database configuration from environment
	logger.AppLogInfo("Loading database configuration")
	dbConfig, err := config.LoadDatabaseConfigFromEnv()
	if err != nil {
		logger.AppLogError("Failed to load database configuration", zap.Error(err))
		exitCode = 1
		return
	}

	// Load Redis configuration from environment
	logger.AppLogInfo("Loading Redis configuration")
	redisConfig, err := config.LoadRedisConfigFromEnv()
	if err != nil {
		logger.AppLogError("Failed to load Redis configuration", zap.Error(err))
		exitCode = 1
		return
	}

	// Setup signal context - cancels on sigterm or sigint
	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Setup ongoing context for inflight requests
	// 	not canceled immediately on shutdown signal
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	defer stopOngoingGracefully()

	// Create pool connection to db with configured timeouts
	logger.AppLogInfo("Connecting to PostgreSQL")
	pool, err := dbConfig.CreatePool(context.Background())
	if err != nil {
		logger.AppLogError("Unable to create connection pool", zap.Error(err))
		exitCode = 1
		return
	}
	defer func() {
		logger.AppLogInfo("Closing PostgreSQL connection pool")
		pool.Close()
		logger.AppLogInfo("PostgreSQL connection pool closed")
	}()
	logger.AppLogInfo("Successfully connected to PostgreSQL")

	// Create Redis client with configured timeouts
	logger.AppLogInfo("Connecting to Redis")
	redisClient, err := redisConfig.CreateClient(context.Background())
	if err != nil {
		logger.AppLogError("Unable to connect to Redis", zap.Error(err))
		exitCode = 1
		return
	}
	defer func() {
		logger.AppLogInfo("Closing Redis connection")
		if err := redisClient.Close(); err != nil {
			logger.AppLogError("Error closing Redis connection", zap.Error(err))
		} else {
			logger.AppLogInfo("Redis connection closed")
		}
	}()
	logger.AppLogInfo("Successfully connected to Redis")

	// Setup DI containers
	postgresRepo := postgres.NewURLRepository(pool, dbConfig.QueryTimeout())

	urlCache := cache_redis.NewURLCache(redisClient)
	urlRepo := repository.NewCachingRepository(postgresRepo, urlCache)
	urlService := service.NewURLService(urlRepo)

	// Setup chi router
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(api_middleware.Logging())
	router.Use(middleware.Timeout(60 * time.Second))

	// Register versioned API routes
	apiConfig := &apiv1.Config{
		URLService:   urlService,
		PgPool:       pool,
		RedisClient:  redisClient,
		ShuttingDown: &isShuttingDown,
	}
	apiv1.RegisterRoutes(router, apiConfig)

	// HTTP Server configuration
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9091"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,

		BaseContext: func(listener net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.AppLogError("Unable to start server", zap.Error(err))
			exitCode = 1
			stop()
		}
		logger.AppLogInfo("Server listening on port", zap.String("port", port))
	}()

	// Wait for shutdown signal
	<-rootCtx.Done()
	stop()

	logger.AppLogInfo("Received shutdown signal")

	// readiness prob should start to return 503
	logger.AppLogInfo("Marking service as unavailable")
	isShuttingDown.Store(true)

	// give time for load balancer to notice service is unavailable
	logger.AppLogInfo(
		"Waiting for load balancer to drain traffic",
		zap.Duration("duration", readinessDrainDelay),
	)
	time.Sleep(readinessDrainDelay)

	// graceful shutdown
	logger.AppLogInfo(
		"Shutting down server",
		zap.Duration("duration", shutdownPeriod),
	)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownPeriod)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.AppLogWarn("Graceful shutdown failed", zap.Error(err))
		logger.AppLogInfo("Cancelling ongoing requests")

		// cancel ongoing requests
		stopOngoingGracefully()

		// hard timeout
		logger.AppLogInfo(
			"Waiting for forced cancellation",
			zap.Duration("duration", shutdownHardPeriod),
		)
		time.Sleep(shutdownHardPeriod)

		logger.AppLogWarn("Forcing server to close")
		if err := server.Close(); err != nil {
			logger.AppLogError("Error forcing server close", zap.Error(err))
			exitCode = 1
		}

	} else {
		logger.AppLogInfo("Server stopped gracefully")
	}

	logger.AppLogInfo("Application shutdown complete")
}
