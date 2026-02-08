package service

import (
	"context"
	"fmt"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/repository"

	"go.uber.org/zap"
)

type UserService interface {
	// GetByID gets user by ID
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByProvider gets user by provider
	GetByProvider(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	logger *zap.Logger,
) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidUser
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByProvider(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error) {
	if !provider.IsValid() {
		return nil, domain.ErrInvalidProvider
	}

	if providerID == "" {
		return nil, domain.ErrInvalidUser
	}

	user, err := s.userRepo.GetByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by provider: %w", err)
	}

	return user, nil
}
