package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"urls_iam_service/internal/domain"
	"urls_iam_service/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) repository.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
  		INSERT INTO sessions (session_id, user_id, created_at, expires_at, last_activity)
  		VALUES ($1, $2, $3, $4, $5)
  	`

	_, err := r.db.Exec(ctx, query,
		session.SessionID,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
  		SELECT session_id, user_id, created_at, expires_at, last_activity
  		FROM sessions
  		WHERE session_id = $1
  	`

	var session domain.Session

	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil

}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	query := `
  		SELECT session_id, user_id, created_at, expires_at, last_activity
  		FROM sessions
  		WHERE user_id = $1
  		ORDER BY created_at DESC
  	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user_id: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session

	for rows.Next() {
		var session domain.Session
		if err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.LastActivity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return sessions, nil
}

func (r *sessionRepository) UpdateLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error {
	query := `
  		UPDATE sessions
  		SET last_activity = $2
  		WHERE session_id = $1
  	`

	result, err := r.db.Exec(ctx, query, sessionID, lastActivity)
	if err != nil {
		return fmt.Errorf("failed to update last_activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) UpdateExpiresAt(ctx context.Context, sessionID string, expiresAt time.Time) error {
	query := `
  		UPDATE sessions
  		SET expires_at = $2
  		WHERE session_id = $1
  	`

	result, err := r.db.Exec(ctx, query, sessionID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to update expires_at: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user_id: %w", err)
	}

	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return result.RowsAffected(), nil
}
