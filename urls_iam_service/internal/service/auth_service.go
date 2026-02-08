package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/repository"
	"urls_iam_service/internal/repository/redis"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService interface {
	// ValidateSession validates existing session (for ForwardAuth)
	ValidateSession(ctx context.Context, sessionID string) (*SessionInfo, error)
	// CreateOrValidateSession creates new session or validates existing one
	CreateOrValidateSession(ctx context.Context, sessionID string, metadata map[string]string) (*SessionResult, error)
	// RefreshSession refreshes session TTL
	RefreshSession(ctx context.Context, sessionID string) (time.Time, error)
	// Logout deletes session
	Logout(ctx context.Context, sessionID string) error
}

type SessionInfo struct {
	UserID   string
	Provider domain.Provider
}

type SessionResult struct {
	UserID    string
	SessionID string
	IsNewUser bool
	ExpiresAt time.Time
	Provider  domain.Provider
}

type authService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	sessionCache *redis.SessionCache
	sidGenerator *SIDGenerator
	sessionTTL   time.Duration
	logger       *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	sessionCache *redis.SessionCache,
	sidGenerator *SIDGenerator,
	sessionTTL time.Duration,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		sessionCache: sessionCache,
		sidGenerator: sidGenerator,
		sessionTTL:   sessionTTL,
		logger:       logger,
	}
}

func (s *authService) ValidateSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	if sessionID == "" {
		return nil, domain.ErrInvalidSession
	}

	// 1. Try to get from cache (fast path)
	session, err := s.sessionCache.Get(ctx, sessionID)
	if err == nil {
		// Cache hit
		if session.ExpiresAt.Before(time.Now()) {
			return nil, domain.ErrSessionExpired
		}

		// Get user info
		user, err := s.userRepo.GetByID(ctx, session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Async update last_activity (don't block response)
		go s.updateLastActivityAsync(sessionID)

		return &SessionInfo{
			UserID:   user.ID,
			Provider: user.Provider,
		}, nil
	}

	// 2. Cache miss - get from database
	session, err = s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 3. Check expiration
	if session.ExpiresAt.Before(time.Now()) {
		// Delete expired session
		go s.deleteSessionAsync(sessionID)
		return nil, domain.ErrSessionExpired
	}

	// 4. Get user info
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 5. Cache the session
	if err := s.sessionCache.Set(ctx, session); err != nil {
		s.logger.Warn("failed to cache session", zap.Error(err))
	}

	// 6. Async update last_activity
	go s.updateLastActivityAsync(sessionID)

	return &SessionInfo{
		UserID:   user.ID,
		Provider: user.Provider,
	}, nil
}

func (s *authService) CreateOrValidateSession(ctx context.Context, sessionID string, metadata map[string]string) (*SessionResult, error) {
	// If sessionID provided, try to validate existing session
	if sessionID != "" {
		sessionInfo, err := s.ValidateSession(ctx, sessionID)
		if err == nil {
			// Valid session exists
			session, _ := s.sessionRepo.GetByID(ctx, sessionID)
			return &SessionResult{
				UserID:    sessionInfo.UserID,
				SessionID: sessionID,
				IsNewUser: false,
				ExpiresAt: session.ExpiresAt,
				Provider:  sessionInfo.Provider,
			}, nil
		}

		// Session invalid or expired - create new one
		s.logger.Debug("session invalid, creating new one",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	// Create new anonymous user and session
	return s.createNewSession(ctx, metadata)
}

func (s *authService) createNewSession(ctx context.Context, metadata map[string]string) (*SessionResult, error) {
	// 1. Generate anonymous provider_id
	providerID, err := s.sidGenerator.GenerateAnonymousProviderID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate provider_id: %w", err)
	}

	// 2. Create user
	now := time.Now()
	user := &domain.User{
		ID:         uuid.New().String(),
		Provider:   domain.ProviderSID,
		ProviderID: providerID,
		Metadata:   metadata,
		CreatedAt:  now,
		LastSeenAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 3. Generate session_id
	sessionID, err := s.sidGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session_id: %w", err)
	}

	// 4. Create session
	expiresAt := now.Add(s.sessionTTL)
	session := &domain.Session{
		SessionID:    sessionID,
		UserID:       user.ID,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastActivity: now,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 5. Cache session
	if err := s.sessionCache.Set(ctx, session); err != nil {
		s.logger.Warn("failed to cache new session", zap.Error(err))
	}

	s.logger.Info("new user and session created",
		zap.String("user_id", user.ID),
		zap.String("session_id", sessionID),
		zap.String("provider", string(user.Provider)),
	)

	return &SessionResult{
		UserID:    user.ID,
		SessionID: sessionID,
		IsNewUser: true,
		ExpiresAt: expiresAt,
		Provider:  user.Provider,
	}, nil
}

func (s *authService) RefreshSession(ctx context.Context, sessionID string) (time.Time, error) {
	if sessionID == "" {
		return time.Time{}, domain.ErrInvalidSession
	}

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return time.Time{}, err
	}

	// Check if expired
	if session.ExpiresAt.Before(time.Now()) {
		return time.Time{}, domain.ErrSessionExpired
	}

	// Update expires_at
	newExpiresAt := time.Now().Add(s.sessionTTL)
	if err := s.sessionRepo.UpdateExpiresAt(ctx, sessionID, newExpiresAt); err != nil {
		return time.Time{}, fmt.Errorf("failed to update expires_at: %w", err)
	}

	// Invalidate cache
	if err := s.sessionCache.Delete(ctx, sessionID); err != nil {
		s.logger.Warn("failed to invalidate cache", zap.Error(err))
	}

	s.logger.Info("session refreshed",
		zap.String("session_id", sessionID),
		zap.Time("new_expires_at", newExpiresAt),
	)

	return newExpiresAt, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return domain.ErrInvalidSession
	}

	// Delete from database
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete from cache
	if err := s.sessionCache.Delete(ctx, sessionID); err != nil {
		s.logger.Warn("failed to delete session from cache", zap.Error(err))
	}

	s.logger.Info("user logged out", zap.String("session_id", sessionID))

	return nil
}

// Async helpers (don't block main request)
func (s *authService) updateLastActivityAsync(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.sessionRepo.UpdateLastActivity(ctx, sessionID, time.Now()); err != nil {
		s.logger.Warn("failed to update last_activity",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}
}

func (s *authService) deleteSessionAsync(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		s.logger.Warn("failed to delete expired session",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	// Also delete from cache
	if err := s.sessionCache.Delete(ctx, sessionID); err != nil {
		s.logger.Warn("failed to delete session from cache",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}
}
