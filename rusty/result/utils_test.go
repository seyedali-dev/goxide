// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package result_test. utils_test provides utilities for testing.
package result_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/seyedali-dev/goxide/rusty/result"
)

// User represents a simple database entity
type User struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// UserRepository using traditional error handling
type TraditionalUserRepo struct {
	db *sql.DB
}

func NewTraditionalUserRepo(db *sql.DB) *TraditionalUserRepo {
	return &TraditionalUserRepo{db: db}
}

func (r *TraditionalUserRepo) CreateUser(ctx context.Context, email, name string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (email, name, created_at) VALUES ($1, $2, $3) RETURNING id",
		email, name, time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (r *TraditionalUserRepo) FindUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (r *TraditionalUserRepo) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err // Traditional approach often returns nil for not found
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *TraditionalUserRepo) UpdateUserName(ctx context.Context, id int, name string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE users SET name = $1 WHERE id = $2",
		name, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UserRepository using Result pattern
type ResultUserRepo struct {
	db *sql.DB
}

func NewResultUserRepo(db *sql.DB) *ResultUserRepo {
	return &ResultUserRepo{db: db}
}

func (r *ResultUserRepo) CreateUser(ctx context.Context, email, name string) result.Result[int] {
	var id int
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (email, name, created_at) VALUES ($1, $2, $3) RETURNING id",
		email, name, time.Now(),
	).Scan(&id)
	return result.Wrap(id, err)
}

func (r *ResultUserRepo) FindUserByID(ctx context.Context, id int) result.Result[*User] {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result.Err[*User](fmt.Errorf("user not found: %w", err))
		}
		return result.Err[*User](fmt.Errorf("failed to find user: %w", err))
	}
	return result.Ok(&user)
}

func (r *ResultUserRepo) FindUserByEmail(ctx context.Context, email string) result.Result[*User] {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result.Err[*User](fmt.Errorf("user not found: %w", err))
		}
		return result.Err[*User](fmt.Errorf("failed to find user by email: %w", err))
	}
	return result.Ok(&user)
}

func (r *ResultUserRepo) UpdateUserName(ctx context.Context, id int, name string) result.Result[bool] {
	resultExec, err := r.db.ExecContext(ctx,
		"UPDATE users SET name = $1 WHERE id = $2",
		name, id,
	)
	if err != nil {
		return result.Err[bool](fmt.Errorf("failed to update user: %w", err))
	}
	rows, err := resultExec.RowsAffected()
	if err != nil {
		return result.Err[bool](fmt.Errorf("failed to get rows affected: %w", err))
	}
	return result.Ok(rows > 0)
}

// Complex operation: Get or create user with email
func (r *TraditionalUserRepo) GetOrCreateUser(ctx context.Context, email, name string) (*User, error) {
	// Try to find existing user
	user, err := r.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Create new user
	id, err := r.CreateUser(ctx, email, name)
	if err != nil {
		return nil, err
	}

	// Return the newly created user
	return r.FindUserByID(ctx, id)
}

func (r *ResultUserRepo) GetOrCreateUser(ctx context.Context, email, name string) result.Result[*User] {
	// Using BubbleUp for early returns with Catch
	var res result.Result[*User]
	defer result.Catch(&res)

	// Try to find existing user first
	userResult := r.FindUserByEmail(ctx, email)
	if userResult.IsOk() {
		return userResult
	}

	// If not found, create new user
	id := r.CreateUser(ctx, email, name).BubbleUp()

	// Return the newly created user
	return r.FindUserByID(ctx, id)
}
