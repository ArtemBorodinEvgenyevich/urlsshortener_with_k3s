package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/repository"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type urlRepository struct {
	psql         sq.StatementBuilderType
	connPool     *pgxpool.Pool
	queryTimeout time.Duration
}

func NewURLRepository(connPool *pgxpool.Pool, queryTimeout time.Duration) repository.URLRepository {
	return &urlRepository{
		connPool:     connPool,
		queryTimeout: queryTimeout,
		psql:         sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (repo *urlRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	// Add timeout for query execution
	ctx, cancel := context.WithTimeout(ctx, repo.queryTimeout)
	defer cancel()

	query, args, err := repo.psql.
		Select("short_code", "original_url", "user_id", "expires_at", "created_at").
		From("urls").
		Where(sq.Eq{"short_code": shortCode}).
		Where(sq.Gt{"expires_at": time.Now()}).
		ToSql()
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't build query", zap.Error(err))
		return nil, err
	}
	logger.PgLogInfo("Query:", zap.String("query", query), zap.Any("args", args))

	url := &domain.URL{}

	err = repo.connPool.QueryRow(ctx, query, args...).Scan(
		&url.ShortCode,
		&url.OriginalURL,
		&url.UserID,
		&url.ExpiresAt,
		&url.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			logger.PgLogInfoCtx(ctx, "", zap.Error(err))
			return nil, repository.ErrNotFound
		default:
			logger.PgLogErrorCtx(ctx, "Can't execute query", zap.Error(err))
			return nil, err
		}
	}

	return url, nil
}

func (repo *urlRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.queryTimeout)
	defer cancel()

	query, args, err := repo.psql.
		Select("short_code", "original_url", "user_id", "expires_at", "created_at").
		From("urls").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Gt{"expires_at": time.Now()}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't build query", zap.Error(err))
		return nil, err
	}
	logger.PgLogInfo("Query:", zap.String("query", query), zap.Any("args", args))

	rows, err := repo.connPool.Query(ctx, query, args...)
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't execute query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var urls []*domain.URL
	for rows.Next() {
		url := &domain.URL{}
		if err := rows.Scan(&url.ShortCode, &url.OriginalURL, &url.UserID, &url.ExpiresAt, &url.CreatedAt); err != nil {
			logger.PgLogErrorCtx(ctx, "Can't scan row", zap.Error(err))
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		logger.PgLogErrorCtx(ctx, "Rows error", zap.Error(err))
		return nil, err
	}

	return urls, nil
}

func (repo *urlRepository) Delete(ctx context.Context, shortCode string) error {
	// Add timeout for query execution
	ctx, cancel := context.WithTimeout(ctx, repo.queryTimeout)
	defer cancel()

	query, args, err := repo.psql.
		Delete("urls").
		Where(sq.Eq{"short_code": shortCode}).
		ToSql()
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't build query", zap.Error(err))
		return err
	}
	logger.PgLogInfo("Query:", zap.String("query", query), zap.Any("args", args))

	res, err := repo.connPool.Exec(ctx, query, args...)
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't execute query", zap.Error(err))
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (repo *urlRepository) Create(ctx context.Context, url *domain.URL) error {
	// Add timeout for query execution
	ctx, cancel := context.WithTimeout(ctx, repo.queryTimeout)
	defer cancel()

	query, args, err := repo.psql.
		Insert("urls").
		Columns("short_code", "original_url", "user_id", "expires_at", "created_at").
		Values(url.ShortCode, url.OriginalURL, url.UserID, url.ExpiresAt, url.CreatedAt).
		ToSql()
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't build query", zap.Error(err))
		return err
	}
	logger.PgLogInfo("Query:", zap.String("query", query), zap.Any("args", args))

	_, err = repo.connPool.Exec(ctx, query, args...)
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't execute query", zap.Error(err))
	}

	return err
}

func (repo *urlRepository) DeleteByShortCodeAndUserID(ctx context.Context, shortCode string, userID string) error {
	// Add timeout for query execution
	ctx, cancel := context.WithTimeout(ctx, repo.queryTimeout)
	defer cancel()

	query, args, err := repo.psql.
		Delete("urls").
		Where(sq.Eq{"short_code": shortCode}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't build query", zap.Error(err))
		return err
	}
	logger.PgLogInfo("Query:", zap.String("query", query), zap.Any("args", args))

	res, err := repo.connPool.Exec(ctx, query, args...)
	if err != nil {
		logger.PgLogErrorCtx(ctx, "Can't execute query", zap.Error(err))
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrForbidden
	}

	return nil
}
