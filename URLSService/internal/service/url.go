package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const maxRetries = 3

var (
	ErrNotFound         = errors.New("URL not found")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidTTL       = errors.New("invalid TTL")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrForbidden        = errors.New("access denied")
)

type URLService interface {
	CreateShortURL(ctx context.Context, originalUrl string, ttlMinutes int, userID *string) (*domain.URL, error)
	GetURL(ctx context.Context, shortCode string) (*domain.URL, error)
	GetUserURLs(ctx context.Context, userID string, limit, offset int) ([]*domain.URL, error)
	DeleteURL(ctx context.Context, shortCode string, userID string) error
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) CreateShortURL(ctx context.Context, originalUrl string, ttlMinutes int, userID *string) (*domain.URL, error) {
	if originalUrl == "" {
		return nil, ErrInvalidURL
	}
	if ttlMinutes <= 0 {
		return nil, ErrInvalidTTL
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		createdAt := time.Now()
		expiresAt := createdAt.Add(time.Minute * time.Duration(ttlMinutes))
		shortCode, err := generateShortCode()
		if err != nil {
			return nil, err
		}

		url := &domain.URL{
			ShortCode:   shortCode,
			OriginalURL: originalUrl,
			UserID:      userID,
			ExpiresAt:   expiresAt,
			CreatedAt:   createdAt,
		}

		err = s.repo.Create(ctx, url)
		if err == nil {
			return url, nil
		}

		if isUniqueViolation(err) {
			lastErr = err
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("failed to generate short code after %d attempts: %w", maxRetries, lastErr)
}

func (s *urlService) GetURL(ctx context.Context, shortCode string) (*domain.URL, error) {
	if shortCode == "" {
		return nil, ErrInvalidShortCode
	}

	url, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return url, nil
}

func (s *urlService) GetUserURLs(ctx context.Context, userID string, limit, offset int) ([]*domain.URL, error) {
	if userID == "" {
		return nil, ErrForbidden
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

func (s *urlService) DeleteURL(ctx context.Context, shortCode string, userID string) error {
	if shortCode == "" {
		return ErrInvalidShortCode
	}

	err := s.repo.DeleteByShortCodeAndUserID(ctx, shortCode, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrNotFound
		case errors.Is(err, repository.ErrForbidden):
			return ErrForbidden
		default:
			return err
		}
	}

	return nil
}

func generateShortCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}

	return false
}
