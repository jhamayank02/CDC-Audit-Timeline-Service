package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

var (
	create_user_query = `
		INSERT INTO users (id, first_name, last_name, email, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::UUID, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	update_user_query = `
		UPDATE users SET first_name = COALESCE(NULLIF($2, ''), first_name), last_name = COALESCE(NULLIF($3, ''), last_name), email = COALESCE(NULLIF($4, ''), email), updated_by = COALESCE(NULLIF($5, '')::UUID, updated_by), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	get_user_query = `
		SELECT id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at FROM users WHERE id = $1
	`
	get_users_query = `
		SELECT id, first_name, last_name, email, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at, COUNT(*) OVER() AS total_count FROM users
	`
)

func (r *repository) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) {
	id := uuid.New()

	var user User
	err := r.db.QueryRowContext(ctx, create_user_query, id, req.FirstName, req.LastName, req.Email, req.CreatedBy).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedBy, &user.UpdatedBy, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
		return nil, err
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx context.Context, req *UpdateUserReq) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, update_user_query, req.Id, req.FirstName, req.LastName, req.Email, req.UpdatedBy).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedBy, &user.UpdatedBy, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, get_user_query, id).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedBy, &user.UpdatedBy, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan user", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetUsers(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
	query := fmt.Sprintf("%s ORDER BY %s %s LIMIT %d OFFSET %d", get_users_query, orderBy, sortBy, limit, offset)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]User, 0)
	var totalCount int
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedBy, &user.UpdatedBy, &user.CreatedAt, &user.UpdatedAt, &totalCount)
		if err != nil {
			r.logger.Error("[REPOSITORY] failed to scan user", "err", err)
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, totalCount, nil
}
