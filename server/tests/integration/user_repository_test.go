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

func basicUser(id uuid.UUID) *domain.User {
	return &domain.User{
		ID:           id,
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
}

func TestSave(t *testing.T) {
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
				Role:         domain.RoleMIPYME,
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
				Role:         domain.RoleMIPYME,
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
				Role:         domain.RoleMIPYME,
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
				Role:         domain.RoleMIPYME,
				Address:      domain.Address{},
				PhoneNumber:  "1234-5678",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			Name:        "No Role",
			ExpectedErr: domain.ErrInvalidInput,
			User: &domain.User{
				ID:           testUserID,
				FirstName:    "John",
				LastName:     "Doe",
				Role:         domain.RolePending,
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
				Role:        domain.RoleMIPYME,
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
				Role:         domain.RoleMIPYME,
				Address:      domain.Address{},
				Email:        "john@example.com",
				PasswordHash: "lamejorcontrasenia1233",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
	}

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

func TestFindByID(t *testing.T) {
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

	savedUser := basicUser(testUserID)
	id, err := db.Save(context.Background(), savedUser)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	_ = id

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testUserID,
			ExpectedErr: nil,
		},
		{
			Name:        "User Not Found",
			ID:          testUserID2,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			user, err := db.FindByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByID() unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Fatal("FindByID() returned nil user")
			}
			if user.ID != testUserID {
				t.Errorf("FindByID() ID = %v, want %v", user.ID, testUserID)
			}
			if user.FirstName != savedUser.FirstName {
				t.Errorf("FindByID() FirstName = %v, want %v", user.FirstName, savedUser.FirstName)
			}
			if user.LastName != savedUser.LastName {
				t.Errorf("FindByID() LastName = %v, want %v", user.LastName, savedUser.LastName)
			}
			if user.Email != savedUser.Email {
				t.Errorf("FindByID() Email = %v, want %v", user.Email, savedUser.Email)
			}
			if user.PhoneNumber != savedUser.PhoneNumber {
				t.Errorf("FindByID() PhoneNumber = %v, want %v", user.PhoneNumber, savedUser.PhoneNumber)
			}
			if user.PasswordHash != savedUser.PasswordHash {
				t.Errorf("FindByID() PasswordHash = %v, want %v", user.PasswordHash, savedUser.PasswordHash)
			}
		})
	}
}

func TestFindByEmail(t *testing.T) {
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

	savedUser := basicUser(testUserID)
	_, err = db.Save(context.Background(), savedUser)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		Email       string
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			Email:       "john@example.com",
			ExpectedErr: nil,
		},
		{
			Name:        "User Not Found",
			Email:       "notfound@example.com",
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			user, err := db.FindByEmail(context.Background(), tt.Email)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByEmail() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByEmail() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByEmail() unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Fatal("FindByEmail() returned nil user")
			}
			if user.Email != tt.Email {
				t.Errorf("FindByEmail() Email = %v, want %v", user.Email, tt.Email)
			}
			if user.ID != savedUser.ID {
				t.Errorf("FindByEmail() ID = %v, want %v", user.ID, savedUser.ID)
			}
			if user.FirstName != savedUser.FirstName {
				t.Errorf("FindByEmail() FirstName = %v, want %v", user.FirstName, savedUser.FirstName)
			}
		})
	}
}

func TestExistsByEmail(t *testing.T) {
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

	savedUser := basicUser(testUserID)
	_, err = db.Save(context.Background(), savedUser)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name     string
		Email    string
		Expected bool
	}{
		{
			Name:     "Email exists",
			Email:    "john@example.com",
			Expected: true,
		},
		{
			Name:     "Email does not exist",
			Email:    "notfound@example.com",
			Expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			exists, err := db.ExistsByEmail(context.Background(), tt.Email)
			if err != nil {
				t.Errorf("ExistsByEmail() unexpected error: %v", err)
				return
			}
			if exists != tt.Expected {
				t.Errorf("ExistsByEmail() = %v, want %v", exists, tt.Expected)
			}
		})
	}
}

func TestExistsByID(t *testing.T) {
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

	savedUser := basicUser(testUserID)
	_, err = db.Save(context.Background(), savedUser)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name     string
		ID       string
		Expected bool
	}{
		{
			Name:     "ID exists",
			ID:       testUserID.String(),
			Expected: true,
		},
		{
			Name:     "ID does not exist",
			ID:       testUserID2.String(),
			Expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			exists, err := db.ExistsByID(context.Background(), tt.ID)
			if err != nil {
				t.Errorf("ExistsByID() unexpected error: %v", err)
				return
			}
			if exists != tt.Expected {
				t.Errorf("ExistsByID() = %v, want %v", exists, tt.Expected)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
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

	savedUser := basicUser(testUserID)
	id, err := db.Save(context.Background(), savedUser)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("uuid.Parse() error: %v", err)
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
					ID:           user_ID,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "User not found",
			ExpectedErr: domain.ErrUserNotFound,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           testUserID2,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No First Name",
			ExpectedErr: domain.ErrFirstNameRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					LastName:     "Doe",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No Last Name",
			ExpectedErr: domain.ErrLastNameRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					FirstName:    "Jane",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No Email",
			ExpectedErr: domain.ErrEmailRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No Role",
			ExpectedErr: domain.ErrInvalidInput,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RolePending,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No valid Role",
			ExpectedErr: domain.ErrValidRoleRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RoleOptions(99),
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PhoneNumber:  "8765-4321",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
		{
			Name:        "No password",
			ExpectedErr: domain.ErrPasswordRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:          user_ID,
					FirstName:   "Jane",
					LastName:    "Doe",
					Role:        domain.RoleAdmin,
					Address:     domain.Address{},
					Email:       "jane@example.com",
					PhoneNumber: "8765-4321",
					UpdatedAt:   fixedTime,
				}
			},
		},
		{
			Name:        "No phone number",
			ExpectedErr: domain.ErrPhoneNumberRequired,
			update_params: func(user_ID uuid.UUID) *domain.User {
				return &domain.User{
					ID:           user_ID,
					FirstName:    "Jane",
					LastName:     "Doe",
					Role:         domain.RoleAdmin,
					Address:      domain.Address{},
					Email:        "jane@example.com",
					PasswordHash: "nuevacontrasena456",
					UpdatedAt:    fixedTime,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Update(context.Background(), tt.update_params(parsedID))

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
			}
		})
	}
}
