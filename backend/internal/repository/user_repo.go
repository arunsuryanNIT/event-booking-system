package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/google/uuid"
)

// UserRepo is the PostgreSQL implementation of UserRepository.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo returns a UserRepo backed by the given database connection.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// ListUsers returns all pre-seeded users ordered by name.
func (r *UserRepo) ListUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, created_at FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// GetUserByID returns a single user by primary key, or ErrUserNotFound.
func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
