package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

func TestUserUseCaseRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		email      string
		firstName  string
		lastName   string
		role       domain.RoleOptions
		password   string
		confirm    string
		existsErr  error
		emailTaken bool
		hashErr    error
		saveErr    error
		wantErr    error
	}{
		{name: "happy path", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123"},
		{name: "invalid role", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RolePending, password: "secret123", confirm: "secret123", wantErr: domain.ErrInvalidInput},
		{name: "password mismatch", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleProvider, password: "secret123", confirm: "other456", wantErr: domain.ErrInvalidInput},
		{name: "empty email", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", wantErr: domain.ErrEmailRequired},
		{name: "empty first name", email: "register@milpa.com.ni", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", wantErr: domain.ErrFirstNameRequired},
		{name: "empty last name", email: "register@milpa.com.ni", firstName: "Jane", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", wantErr: domain.ErrLastNameRequired},
		{name: "repo error", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", existsErr: errFake, wantErr: errFake},
		{name: "email taken", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", emailTaken: true, wantErr: domain.ErrEmailTaken},
		{name: "hash error", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", hashErr: errFake, wantErr: errFake},
		{name: "save error", email: "register@milpa.com.ni", firstName: "Jane", lastName: "Smith", role: domain.RoleMIPYME, password: "secret123", confirm: "secret123", saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			if tt.existsErr != nil {
				userRepo.existsByEmail = func(ctx context.Context, email string) (bool, error) {
					return false, tt.existsErr
				}
			}
			if tt.emailTaken {
				userRepo.existsByEmail = func(ctx context.Context, email string) (bool, error) {
					return true, nil
				}
			}
			if tt.saveErr != nil {
				userRepo.save = func(ctx context.Context, user *domain.User) error {
					return tt.saveErr
				}
			}
			hasher := newFakeHasher()
			if tt.hashErr != nil {
				hasher.hash = func(password string) (string, error) {
					return "", tt.hashErr
				}
			}
			uc := usecases.NewUserUseCase(userRepo, hasher, newFakeJWT(), newFakeTimer())

			got, err := uc.Register(context.Background(), dto.RegisterUserRequest{
				Email:           tt.email,
				FirstName:       tt.firstName,
				LastName:        tt.lastName,
				Role:            tt.role,
				Address:         "Managua",
				Password:        tt.password,
				ConfirmPassword: tt.confirm,
				PhoneNumber:     "555-1234",
			})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if got.Email != tt.email {
				t.Errorf("email = %q, want %q", got.Email, tt.email)
			}
			if got.FirstName != tt.firstName {
				t.Errorf("first name = %q, want %q", got.FirstName, tt.firstName)
			}
			if got.LastName != tt.lastName {
				t.Errorf("last name = %q, want %q", got.LastName, tt.lastName)
			}
			if got.Role != tt.role {
				t.Errorf("role = %v, want %v", got.Role, tt.role)
			}
			if got.PhoneNumber != "555-1234" {
				t.Errorf("phone number = %q, want %q", got.PhoneNumber, "555-1234")
			}
			if got.Address != "Managua, , " {
				t.Errorf("address = %q, want %q", got.Address, "Managua, , ")
			}
			if !got.CreatedAt.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", got.CreatedAt, fixedTime)
			}
			if !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", got.UpdatedAt, fixedTime)
			}
			if len(userRepo.existedEmails) != 1 || userRepo.existedEmails[0] != tt.email {
				t.Errorf("existed emails = %v, want [%q]", userRepo.existedEmails, tt.email)
			}
			if len(userRepo.saved) != 1 {
				t.Fatalf("saved users = %d, want 1", len(userRepo.saved))
			}
			saved := userRepo.saved[0]
			if saved.PasswordHash != "hashed-"+tt.password {
				t.Errorf("password hash = %q, want %q", saved.PasswordHash, "hashed-"+tt.password)
			}
			if saved.PhoneNumber != "555-1234" {
				t.Errorf("saved phone number = %q, want %q", saved.PhoneNumber, "555-1234")
			}
			if saved.Address.AddressLine != "Managua" {
				t.Errorf("saved address line = %q, want %q", saved.Address.AddressLine, "Managua")
			}
			if saved.Role != tt.role {
				t.Errorf("saved role = %v, want %v", saved.Role, tt.role)
			}
		})
	}
}

func TestUserUseCaseLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		repoErr    error
		compareErr error
		tokenErr   error
		wantErr    error
	}{
		{name: "happy path"},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
		{name: "not found", repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "wrong password", compareErr: errFake, wantErr: domain.ErrUnauthorized},
		{name: "token error", tokenErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			if tt.repoErr != nil {
				userRepo.findByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return nil, tt.repoErr
				}
			}
			hasher := newFakeHasher()
			if tt.compareErr != nil {
				hasher.compare = func(hash, password string) error {
					return tt.compareErr
				}
			}
			jwt := newFakeJWT()
			if tt.tokenErr != nil {
				jwt.generateToken = func(userID uuid.UUID, role domain.RoleOptions) (string, error) {
					return "", tt.tokenErr
				}
			}
			uc := usecases.NewUserUseCase(userRepo, hasher, jwt, newFakeTimer())

			got, err := uc.Login(context.Background(), dto.LoginRequest{Email: "user@milpa.com.ni", Password: "secret123"})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.AccessToken != "signed-token" {
				t.Errorf("access token = %q, want %q", got.AccessToken, "signed-token")
			}
			if got.ExpiresIn != 86400 {
				t.Errorf("expires in = %d, want 86400", got.ExpiresIn)
			}
			if jwt.tokenUserID != testUserID {
				t.Errorf("token user id = %v, want %v", jwt.tokenUserID, testUserID)
			}
			if jwt.tokenRole != domain.RolePending {
				t.Errorf("token role = %v, want %v", jwt.tokenRole, domain.RolePending)
			}
			if got.User.ID != testUserID {
				t.Errorf("user id = %v, want %v", got.User.ID, testUserID)
			}
			if got.User.Email != "user@milpa.com.ni" {
				t.Errorf("user email = %q, want %q", got.User.Email, "user@milpa.com.ni")
			}
			if got.User.FirstName != "John" || got.User.LastName != "Doe" {
				t.Errorf("user name = %q %q, want John Doe", got.User.FirstName, got.User.LastName)
			}
		})
	}
}

func TestUserUseCaseGetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "happy path"},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
		{name: "not found", repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			if tt.repoErr != nil {
				userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

			got, err := uc.GetByID(context.Background(), testUserID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != testUserID {
				t.Errorf("id = %v, want %v", got.ID, testUserID)
			}
			if got.Email != "user@milpa.com.ni" {
				t.Errorf("email = %q, want %q", got.Email, "user@milpa.com.ni")
			}
			if got.FirstName != "John" || got.LastName != "Doe" {
				t.Errorf("name = %q %q, want John Doe", got.FirstName, got.LastName)
			}
			if got.Role != domain.RolePending {
				t.Errorf("role = %v, want %v", got.Role, domain.RolePending)
			}
			if got.Address != ", , " {
				t.Errorf("address = %q, want %q", got.Address, ", , ")
			}
			if got.PhoneNumber != "" {
				t.Errorf("phone number = %q, want empty", got.PhoneNumber)
			}
			if !got.CreatedAt.Equal(fixedTime) || !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestUserUseCaseUpdateProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		req             dto.UpdateUserRequest
		repoErr         error
		wantErr         error
		wantEmail       string
		wantFirstName   string
		wantLastName    string
		wantAddressLine string
		wantPhone       string
	}{
		{
			name:          "no fields",
			req:           dto.UpdateUserRequest{},
			wantEmail:     "user@milpa.com.ni",
			wantFirstName: "John",
			wantLastName:  "Doe",
		},
		{
			name: "all fields",
			req: dto.UpdateUserRequest{
				Email:       strPtr("new@milpa.com.ni"),
				FirstName:   strPtr("Jane"),
				LastName:    strPtr("Roe"),
				Address:     strPtr("Managua"),
				PhoneNumber: strPtr("888-0000"),
			},
			wantEmail:       "new@milpa.com.ni",
			wantFirstName:   "Jane",
			wantLastName:    "Roe",
			wantAddressLine: "Managua",
			wantPhone:       "888-0000",
		},
		{name: "email only", req: dto.UpdateUserRequest{Email: strPtr("new@milpa.com.ni")}, wantEmail: "new@milpa.com.ni", wantFirstName: "John", wantLastName: "Doe"},
		{name: "first name only", req: dto.UpdateUserRequest{FirstName: strPtr("Jane")}, wantEmail: "user@milpa.com.ni", wantFirstName: "Jane", wantLastName: "Doe"},
		{name: "last name only", req: dto.UpdateUserRequest{LastName: strPtr("Roe")}, wantEmail: "user@milpa.com.ni", wantFirstName: "John", wantLastName: "Roe"},
		{name: "address only", req: dto.UpdateUserRequest{Address: strPtr("Managua")}, wantEmail: "user@milpa.com.ni", wantFirstName: "John", wantLastName: "Doe", wantAddressLine: "Managua"},
		{name: "phone only", req: dto.UpdateUserRequest{PhoneNumber: strPtr("888-0000")}, wantEmail: "user@milpa.com.ni", wantFirstName: "John", wantLastName: "Doe", wantPhone: "888-0000"},
		{name: "repo error", req: dto.UpdateUserRequest{FirstName: strPtr("Jane")}, repoErr: errFake, wantErr: errFake},
		{name: "not found", req: dto.UpdateUserRequest{FirstName: strPtr("Jane")}, repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			if tt.repoErr != nil {
				userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

			err := uc.UpdateProfile(context.Background(), testUserID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(userRepo.updated) != 1 {
				t.Fatalf("updated users = %d, want 1", len(userRepo.updated))
			}
			updated := userRepo.updated[0]
			if updated.Email != tt.wantEmail {
				t.Errorf("email = %q, want %q", updated.Email, tt.wantEmail)
			}
			if updated.FirstName != tt.wantFirstName {
				t.Errorf("first name = %q, want %q", updated.FirstName, tt.wantFirstName)
			}
			if updated.LastName != tt.wantLastName {
				t.Errorf("last name = %q, want %q", updated.LastName, tt.wantLastName)
			}
			if updated.Address.AddressLine != tt.wantAddressLine {
				t.Errorf("address line = %q, want %q", updated.Address.AddressLine, tt.wantAddressLine)
			}
			if updated.PhoneNumber != tt.wantPhone {
				t.Errorf("phone number = %q, want %q", updated.PhoneNumber, tt.wantPhone)
			}
			if !updated.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", updated.UpdatedAt, fixedTime)
			}
		})
	}
}
