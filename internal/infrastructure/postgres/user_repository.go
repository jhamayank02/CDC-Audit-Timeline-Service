package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

var (
	createUserQuery = `
		INSERT INTO users (id, first_name, last_name, email, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::UUID, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	updateUserQuery = `
		UPDATE users SET first_name = COALESCE(NULLIF($2, ''), first_name), last_name = COALESCE(NULLIF($3, ''), last_name), email = COALESCE(NULLIF($4, ''), email), updated_by = COALESCE(NULLIF($5, '')::UUID, updated_by), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	getUserQuery = `
		SELECT id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at FROM users WHERE id = $1
	`
	listUsersQuery = `
		SELECT id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at, COUNT(*) OVER() AS total_count FROM users
	`
)

func (r *UserRepository) Create(ctx context.Context, input user.CreateInput) (*user.User, error) {
	id := uuid.New()

	var result user.User
	err := r.db.QueryRowContext(ctx, createUserQuery, id, input.FirstName, input.LastName, input.Email, input.CreatedBy).
		Scan(&result.ID, &result.FirstName, &result.LastName, &result.Email, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
		return nil, err
	}

	return &result, nil
}

func (r *UserRepository) Update(ctx context.Context, input user.UpdateInput) (*user.User, error) {
	var result user.User
	err := r.db.QueryRowContext(ctx, updateUserQuery, input.ID, input.FirstName, input.LastName, input.Email, input.UpdatedBy).
		Scan(&result.ID, &result.FirstName, &result.LastName, &result.Email, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var result user.User
	err := r.db.QueryRowContext(ctx, getUserQuery, id).
		Scan(&result.ID, &result.FirstName, &result.LastName, &result.Email, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]user.User, int, error) {
	query := fmt.Sprintf("%s ORDER BY %s %s LIMIT %d OFFSET %d", listUsersQuery, orderBy, sortBy, limit, offset)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]user.User, 0)
	var totalCount int
	for rows.Next() {
		var result user.User
		err := rows.Scan(&result.ID, &result.FirstName, &result.LastName, &result.Email, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt, &totalCount)
		if err != nil {
			r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
			return nil, 0, err
		}
		users = append(users, result)
	}

	return users, totalCount, rows.Err()
}
