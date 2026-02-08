package service

import (
	"context"
	"time"
	"urls_iam_service/internal/repository"

	"go.uber.org/zap"
)

type CleanupService struct {
	sessionRepo repository.SessionRepository
	logger      *zap.Logger
	interval    time.Duration
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(
	sessionRepo repository.SessionRepository,
	logger *zap.Logger,
	interval time.Duration,
) *CleanupService {
	if interval == 0 {
		interval = 1 * time.Hour // Default: cleanup every hour
	}

	return &CleanupService{
		sessionRepo: sessionRepo,
		logger:      logger,
		interval:    interval,
	}
}

// Start starts the cleanup job
func (s *CleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("cleanup service started", zap.Duration("interval", s.interval))

	// Run immediately on start
	s.cleanup(ctx)

	for {
		select {
		case <-ticker.C:
			s.cleanup(ctx)
		case <-ctx.Done():
			s.logger.Info("cleanup service stopped")
			return
		}
	}
}

func (s *CleanupService) cleanup(ctx context.Context) {
	s.logger.Debug("starting cleanup of expired sessions")

	deleted, err := s.sessionRepo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("failed to cleanup expired sessions", zap.Error(err))
		return
	}

	if deleted > 0 {
		s.logger.Info("expired sessions cleaned up", zap.Int64("deleted", deleted))
	} else {
		s.logger.Debug("no expired sessions to cleanup")
	}
}
