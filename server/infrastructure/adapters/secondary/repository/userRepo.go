package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type UserRepositoryImpl struct {
	pool DB
}

func NewUserRepository(pool DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{pool: pool}
}

func (userRepo *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {

	var user domain.User

	query := `
		SELECT id, first_name, last_name, role, created_at, updated_at, address_id, email, phone_number, password_hash FROM users WHERE id = $1

	`

	err := userRepo.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.Address.ID, &user.Email, &user.PhoneNumber,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("user.FindByID: %w", err)
	}

	return &user, nil
}

func (userRepo *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	query := `
		SELECT id, first_name, last_name, role, created_at, updated_at, address_id, email, phone_number, password_hash FROM users WHERE email = $1::text
	`

	err := userRepo.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.Address.ID, &user.Email,
		&user.PhoneNumber, &user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user.FindByEmail: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("user.FindByEmail: %w", err)
	}

	return &user, nil
}

func (userRepo *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1::text)
	`

	err := userRepo.pool.QueryRow(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("user.ExistsByEmail: %w", err)
	}

	return exists, nil

}

func nullUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func (userRepo *UserRepositoryImpl) Save(ctx context.Context, user *domain.User) error {

	tx, err := userRepo.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("user.Save: %w", err)
	}

	defer tx.Rollback(ctx)

	var AddressID uuid.UUID

	if user.Address.Department != "" {
		query := `
			INSERT INTO	addresses (id, department, municipality, address_line, latitude, longitude) 
			VALUES ($1, $2, $3, $4, $5, $6)
			returning id
		`

		err = tx.QueryRow(ctx, query, uuid.New(), user.Address.Department, user.Address.Municipality,
			user.Address.AddressLine, user.Address.Latitude, user.Address.Longitude).Scan(&AddressID)

		if err != nil {
			return fmt.Errorf("user.Save: %w", err)
		}

	}

	query := `
		INSERT INTO users (id, first_name, last_name, role, created_at, updated_at, address_id, email, phone_number, password_hash)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.Exec(ctx, query, user.ID, user.FirstName, user.LastName, user.Role, user.CreatedAt, user.UpdatedAt, nullUUID(AddressID), user.Email, user.PhoneNumber,
		user.PasswordHash)

	if err != nil {
		return fmt.Errorf("user.Save: insert user: %v", err)
	}

	return tx.Commit(ctx)
}

func (userRepo *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {

	tx, err := userRepo.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("user.Update: %w", err)
	}

	defer tx.Rollback(ctx)

	if user.Address.ID != uuid.Nil {
		query := ` 
			UPDATE addresses
			SET department = $1, municipality = $2, address_line = $3, latitude = $4, longitude = $5
			WHERE id = $6
		`

		_, err = tx.Exec(ctx, query, user.Address.Department, user.Address.Municipality, user.Address.AddressLine, user.Address.Latitude,
			user.Address.Longitude, user.Address.ID)

		if err != nil {
			return fmt.Errorf("user.Update: address update: %w", err)
		}

	}

	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, role = $3,  updated_at = $4, address_id = $5, email = $6, phone_number = $7, password_hash = $8
		WHERE id = $9
	`

	_, err = tx.Exec(ctx, query, user.FirstName, user.LastName, user.Role, user.UpdatedAt, nullUUID(user.Address.ID), user.Email,
		user.PhoneNumber, user.PasswordHash, user.ID)

	if err != nil {
		return fmt.Errorf("user.Update: %w", err)
	}

	return tx.Commit(ctx)
}

var _ port.UserRepository = (*UserRepositoryImpl)(nil)
