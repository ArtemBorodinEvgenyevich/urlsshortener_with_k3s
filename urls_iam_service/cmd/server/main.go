package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urls_iam_service/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"urls_iam_service/internal/config"
	"urls_iam_service/internal/handler"
	"urls_iam_service/internal/pkg/logger"
	"urls_iam_service/internal/repository/postgres"
	pkgRedis "urls_iam_service/internal/repository/redis"
	"urls_iam_service/internal/service"
)

func main() {
	app := fx.New(
		// Provide configuration
		fx.Provide(config.Load),

		// Provide logger
		fx.Provide(func(cfg *config.Config) (*zap.Logger, error) {
			return logger.New(cfg.App.Env, cfg.App.LogLevel)
		}),

		// Provide database connection
		fx.Provide(func(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
			pool, err := postgres.NewPool(context.Background(), cfg.Postgres, logger)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.Info("closing PostgreSQL connection pool")
					pool.Close()
					return nil
				},
			})

			return pool, nil
		}),

		// Provide Redis client
		fx.Provide(func(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
			client, err := pkgRedis.NewClient(context.Background(), cfg.Redis, logger)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.Info("closing Redis client")
					return client.Close()
				},
			})

			return client, nil
		}),

		// Provide session cache
		fx.Provide(func(client *redis.Client, cfg *config.Config) *pkgRedis.SessionCache {
			return pkgRedis.NewSessionCache(client, cfg.Session.CacheTTL)
		}),

		// Provide repositories (return interfaces!)
		fx.Provide(func(db *pgxpool.Pool) repository.UserRepository {
			return postgres.NewUserRepository(db)
		}),
		fx.Provide(func(db *pgxpool.Pool) repository.SessionRepository {
			return postgres.NewSessionRepository(db)
		}),

		// Provide SID generator
		fx.Provide(func(cfg *config.Config) *service.SIDGenerator {
			return service.NewSIDGenerator(cfg.Session.IDLength)
		}),

		// Provide services (use interfaces in parameters!)
		fx.Provide(func(
			userRepo repository.UserRepository, // ← интерфейс
			sessionRepo repository.SessionRepository, // ← интерфейс
			sessionCache *pkgRedis.SessionCache,
			sidGenerator *service.SIDGenerator,
			cfg *config.Config,
			logger *zap.Logger,
		) service.AuthService {
			return service.NewAuthService(
				userRepo,
				sessionRepo,
				sessionCache,
				sidGenerator,
				cfg.Session.TTL,
				logger,
			)
		}),

		fx.Provide(func(
			userRepo repository.UserRepository, // ← интерфейс
			logger *zap.Logger,
		) service.UserService {
			return service.NewUserService(userRepo, logger)
		}),

		// Provide cleanup service
		fx.Provide(func(
			sessionRepo repository.SessionRepository, // ← интерфейс
			logger *zap.Logger,
		) *service.CleanupService {
			return service.NewCleanupService(sessionRepo, logger, 1*time.Hour)
		}),

		// Provide handlers
		fx.Provide(func(
			authService service.AuthService,
			cfg *config.Config,
			logger *zap.Logger,
		) *handler.AuthHandler {
			return handler.NewAuthHandler(authService, cfg.Cookie, logger)
		}),

		fx.Provide(func(
			userService service.UserService,
			logger *zap.Logger,
		) *handler.UserHandler {
			return handler.NewUserHandler(userService, logger)
		}),

		fx.Provide(func(
			db *pgxpool.Pool,
			redisClient *redis.Client,
			logger *zap.Logger,
		) *handler.HealthHandler {
			return handler.NewHealthHandler(db, redisClient, logger)
		}),

		// Provide handlers struct
		fx.Provide(func(
			auth *handler.AuthHandler,
			user *handler.UserHandler,
			health *handler.HealthHandler,
		) *handler.Handlers {
			return &handler.Handlers{
				Auth:   auth,
				User:   user,
				Health: health,
			}
		}),

		// Provide HTTP server
		fx.Provide(func(
			handlers *handler.Handlers,
			cfg *config.Config,
			logger *zap.Logger,
		) *http.Server {
			router := handler.SetupRouter(handlers, logger)

			return &http.Server{
				Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
				Handler:      router,
				ReadTimeout:  cfg.Server.ReadTimeout,
				WriteTimeout: cfg.Server.WriteTimeout,
				IdleTimeout:  cfg.Server.IdleTimeout,
			}
		}),

		// Invoke HTTP server startup
		fx.Invoke(func(
			lc fx.Lifecycle,
			srv *http.Server,
			cfg *config.Config,
			logger *zap.Logger,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("starting HTTP server",
						zap.Int("port", cfg.Server.Port),
						zap.String("env", cfg.App.Env),
					)

					go func() {
						if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							logger.Fatal("HTTP server failed", zap.Error(err))
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("stopping HTTP server")
					return srv.Shutdown(ctx)
				},
			})
		}),

		// Invoke cleanup service startup
		fx.Invoke(func(
			lc fx.Lifecycle,
			cleanupService *service.CleanupService,
			logger *zap.Logger,
		) {
			// Create context for cleanup service
			ctx, cancel := context.WithCancel(context.Background())

			lc.Append(fx.Hook{
				OnStart: func(startCtx context.Context) error {
					logger.Info("starting cleanup service")

					// Start cleanup service in background
					go cleanupService.Start(ctx)

					return nil
				},
				OnStop: func(stopCtx context.Context) error {
					logger.Info("stopping cleanup service")
					cancel() // Cancel context to stop cleanup service
					return nil
				},
			})
		}),
	)

	// Start application
	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application stopped gracefully")
}
