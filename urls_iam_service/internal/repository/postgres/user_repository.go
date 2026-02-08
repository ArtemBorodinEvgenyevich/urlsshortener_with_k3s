package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
  		INSERT INTO users (id, provider, provider_id, metadata, created_at, last_seen_at)
  		VALUES ($1, $2, $3, $4, $5, $6)
  	`

	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		user.ID,
		user.Provider,
		user.ProviderID,
		metadata,
		user.CreatedAt,
		user.LastSeenAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
  		SELECT id, provider, provider_id, metadata, created_at, last_seen_at
  		FROM users
  		WHERE id = $1
  	`

	var user domain.User
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Provider,
		&user.ProviderID,
		&metadataJSON,
		&user.CreatedAt,
		&user.LastSeenAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByProvider(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error) {
	query := `
  		SELECT id, provider, provider_id, metadata, created_at, last_seen_at
  		FROM users
  		WHERE provider = $1 AND provider_id = $2
  	`

	var user domain.User
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, provider, providerID).Scan(
		&user.ID,
		&user.Provider,
		&user.ProviderID,
		&metadataJSON,
		&user.CreatedAt,
		&user.LastSeenAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by provider: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &user, nil
}

func (r *userRepository) UpdateLastSeen(ctx context.Context, id string, lastSeen time.Time) error {
	query := `
  		UPDATE users
  		SET last_seen_at = $2
  		WHERE id = $1
  	`

	result, err := r.db.Exec(ctx, query, id, lastSeen)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
