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
				userRepo.save = func(ctx context.Context, user *domain.User) (string, error) {
					return "", tt.saveErr
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

func TestUserUseCaseRegisterPersistsLocation(t *testing.T) {
	t.Parallel()

	latitude := 12.435010881390852
	longitude := -86.87811141017944

	userRepo := newFakeUserRepo()
	uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

	_, err := uc.Register(context.Background(), dto.RegisterUserRequest{
		Email:           "geo@milpa.com.ni",
		FirstName:       "Jane",
		LastName:        "Smith",
		Role:            domain.RoleProvider,
		Address:         "Costado Sur del Parque Central",
		Department:      "Leon",
		Municipality:    "Leon",
		Latitude:        &latitude,
		Longitude:       &longitude,
		Password:        "secret123",
		ConfirmPassword: "secret123",
		PhoneNumber:     "555-1234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(userRepo.saved) != 1 {
		t.Fatalf("saved users = %d, want 1", len(userRepo.saved))
	}
	saved := userRepo.saved[0]
	if saved.Address.Department != "Leon" {
		t.Errorf("department = %q, want %q", saved.Address.Department, "Leon")
	}
	if saved.Address.Municipality != "Leon" {
		t.Errorf("municipality = %q, want %q", saved.Address.Municipality, "Leon")
	}
	if saved.Address.Latitude != latitude {
		t.Errorf("latitude = %v, want %v", saved.Address.Latitude, latitude)
	}
	if saved.Address.Longitude != longitude {
		t.Errorf("longitude = %v, want %v", saved.Address.Longitude, longitude)
	}
	if !saved.Address.HasCoordinates() {
		t.Error("HasCoordinates() = false, want true for a registration with coordinates")
	}
}

func TestUserUseCaseRegisterRejectsOutOfRangeCoordinates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{name: "latitude above range", latitude: 90.1, longitude: -86.8},
		{name: "latitude below range", latitude: -91, longitude: -86.8},
		{name: "longitude above range", latitude: 12.4, longitude: 180.5},
		{name: "longitude below range", latitude: 12.4, longitude: -181},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			latitude, longitude := tt.latitude, tt.longitude
			userRepo := newFakeUserRepo()
			uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

			_, err := uc.Register(context.Background(), dto.RegisterUserRequest{
				Email:           "geo@milpa.com.ni",
				FirstName:       "Jane",
				LastName:        "Smith",
				Role:            domain.RoleProvider,
				Department:      "Leon",
				Latitude:        &latitude,
				Longitude:       &longitude,
				Password:        "secret123",
				ConfirmPassword: "secret123",
				PhoneNumber:     "555-1234",
			})

			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want %v", err, domain.ErrInvalidInput)
			}
			if len(userRepo.saved) != 0 {
				t.Errorf("saved users = %d, want 0 for a rejected registration", len(userRepo.saved))
			}
		})
	}
}

func TestUserUseCaseUpdateProfileKeepsAddressIdentity(t *testing.T) {
	t.Parallel()

	addressID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	latitude := 13.0913
	longitude := -86.0014

	userRepo := newFakeUserRepo()
	userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		user := mustUser()
		user.Address = domain.Address{
			ID:           addressID,
			Department:   "Jinotega",
			Municipality: "Jinotega",
			AddressLine:  "Barrio Centro",
			Latitude:     12.4,
			Longitude:    -86.8,
		}
		return user, nil
	}

	uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

	// A change to the address line alone must not orphan the address row.
	err := uc.UpdateProfile(principalCtx(), testUserID, dto.UpdateUserRequest{Address: strPtr("Barrio Nuevo")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(userRepo.updated) != 1 {
		t.Fatalf("updated users = %d, want 1", len(userRepo.updated))
	}
	if userRepo.updated[0].Address.ID != addressID {
		t.Errorf("address id = %v, want the existing %v", userRepo.updated[0].Address.ID, addressID)
	}
	if userRepo.updated[0].Address.AddressLine != "Barrio Nuevo" {
		t.Errorf("address line = %q, want %q", userRepo.updated[0].Address.AddressLine, "Barrio Nuevo")
	}
	if userRepo.updated[0].Address.Department != "Jinotega" {
		t.Errorf("department = %q, want the untouched %q", userRepo.updated[0].Address.Department, "Jinotega")
	}
	if userRepo.updated[0].Address.Latitude != 12.4 {
		t.Errorf("latitude = %v, want the untouched 12.4", userRepo.updated[0].Address.Latitude)
	}

	// A change of coordinates must actually change them.
	err = uc.UpdateProfile(principalCtx(), testUserID, dto.UpdateUserRequest{Latitude: &latitude, Longitude: &longitude})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated := userRepo.updated[len(userRepo.updated)-1]
	if updated.Address.Latitude != latitude || updated.Address.Longitude != longitude {
		t.Errorf("coordinates = (%v, %v), want (%v, %v)", updated.Address.Latitude, updated.Address.Longitude, latitude, longitude)
	}
	if updated.Address.ID != addressID {
		t.Errorf("address id = %v, want the existing %v", updated.Address.ID, addressID)
	}
}

func TestUserUseCaseUpdateProfileRejectsOutOfRangeCoordinates(t *testing.T) {
	t.Parallel()

	latitude := 1000.0
	userRepo := newFakeUserRepo()
	uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

	err := uc.UpdateProfile(principalCtx(), testUserID, dto.UpdateUserRequest{Latitude: &latitude})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want %v", err, domain.ErrInvalidInput)
	}
	if len(userRepo.updated) != 0 {
		t.Errorf("updated users = %d, want 0 for rejected coordinates", len(userRepo.updated))
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

			err := uc.UpdateProfile(principalCtx(), testUserID, tt.req)

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
