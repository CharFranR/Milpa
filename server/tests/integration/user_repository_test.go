package integration

import (
	"context"
	"errors"
	"log"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

var fixedTime time.Time = time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
var testUserID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testUserID2 uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111112")

func TestSave(t *testing.T) {

	tests := []struct {
		Name        string
		ExpectedErr error
		User        *domain.User
	}{
		{
			Name: "Happy Path",
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(1),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No First Name",
			ExpectedErr: domain.ErrFirstNameRequired,
			User: &domain.User{
				ID:           testUserID,
				LastName:     "Doe",
				Role:         domain.RoleOptions(1),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Last Name",
			ExpectedErr: domain.ErrLastNameRequired,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				Role:         domain.RoleOptions(1),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Email",
			ExpectedErr: domain.ErrEmailRequired,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(1),
				Address:      domain.Address{},
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Role 1",
			ExpectedErr: domain.ErrInvalidInput,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(0),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Role 1",
			ExpectedErr: domain.ErrInvalidInput,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(0),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Role 2",
			ExpectedErr: domain.ErrInvalidInput,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No valid Role",
			ExpectedErr: domain.ErrValidRoleRequired,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(99),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No password",
			ExpectedErr: domain.ErrPasswordRequired,
			User: &domain.User{
				ID:          testUserID,
				FirstName:   "John",
				LastName:    "Doe",
				Role:        domain.RoleOptions(1),
				Address:     domain.Address{},
				Email:       "john@example.com",
				PhoneNumber: "1234-5678",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			Name:        "No phone number",
			ExpectedErr: domain.ErrPhoneNumberRequired,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RoleOptions(1),
				Address:      domain.Address{},
				Email:        "john@example.com",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
	}

	PoolConnection, container, err := InitTestDB()

	if err != nil {
		t.Fatalf("fail to init db test connection: %v", err)
	}

	db := repository.NewUserRepository(PoolConnection)

	defer func() {
		if err := testcontainers.TerminateContainer(*container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {

			_, err = db.Save(context.Background(), tt.User)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Save() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Save() unexpected error: %v", err)
			}

		})
	}
}

func TestUpdate(t *testing.T) {

	BasicUser := domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	tests := []struct {
		Name          string
		ExpectedErr   error
		update_params func(user_ID uuid.UUID) *domain.User
	}{
		{
			Name: "Happy path",
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:          user_ID,
					FirstName:   "Jane",
					LastName:    "Doe",
					Email:       "jane@milpa.com",
					PhoneNumber: "8765-4321",
				}

			},
		},
		{
			Name:        "not a valid ID",
			ExpectedErr: domain.ErrUserNotFound,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:          testUserID2,
					FirstName:   BasicUser.FirstName,
					LastName:    BasicUser.LastName,
					Email:       BasicUser.Email,
					PhoneNumber: BasicUser.PhoneNumber,
				}

			},
		},
	}

	PoolConnection, container, err := InitTestDB()

	if err != nil {
		t.Fatalf("fail to init db test connection: %v", err)
	}

	db := repository.NewUserRepository(PoolConnection)

	defer func() {
		if err := testcontainers.TerminateContainer(*container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	id, _ := db.Save(context.Background(), &BasicUser)

	parsedID, _ := uuid.Parse(id)

	db.Save(context.Background(), &BasicUser)

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {

			err := db.Update(context.Background(), tt.update_params(parsedID))

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Save() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Save() unexpected error: %v", err)
			}

		})
	}
}
