package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type fakeUserRepo struct {
	user        *domain.User
	findErr     error
	saveErr     error
	updateErr   error
	saveCalls   int
	updateCalls int
	savedUser   *domain.User
	updatedUser *domain.User
}

var _ port.UserRepository = (*fakeUserRepo)(nil)

func (f *fakeUserRepo) FindByID(context.Context, uuid.UUID) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeUserRepo) FindByEmail(context.Context, string) (*domain.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.user == nil {
		return nil, domain.ErrNotFound
	}
	return f.user, nil
}

func (f *fakeUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return f.user != nil, nil
}

func (f *fakeUserRepo) ExistsByID(context.Context, string) (bool, error) {
	return false, nil
}

func (f *fakeUserRepo) Save(_ context.Context, user *domain.User) (string, error) {
	f.saveCalls++
	f.savedUser = user
	if f.saveErr != nil {
		return "", f.saveErr
	}
	return user.ID.String(), nil
}

func (f *fakeUserRepo) Update(_ context.Context, user *domain.User) error {
	f.updateCalls++
	f.updatedUser = user
	return f.updateErr
}

func (f *fakeUserRepo) List(context.Context, int, int) ([]domain.User, int, error) {
	return nil, 0, nil
}

type fakeHasher struct {
	hashCalls int
	hashErr   error
}

var _ port.PasswordHasher = (*fakeHasher)(nil)

func (f *fakeHasher) Hash(password string) (string, error) {
	f.hashCalls++
	if f.hashErr != nil {
		return "", f.hashErr
	}
	return "hashed:" + password, nil
}

func (f *fakeHasher) Compare(string, string) error {
	return nil
}

type fakeClock struct {
	now time.Time
}

var _ port.TimeProvider = fakeClock{}

func (c fakeClock) Now() time.Time {
	return c.now
}

func TestBootstrapEnsureAdmin(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 10, 8, 15, 4, 5, 0, time.UTC)
	existingTime := fixedNow.Add(-48 * time.Hour)
	findFailure := errors.New("database unavailable")

	adminID := uuid.New()
	farmerID := uuid.New()

	tests := []struct {
		name     string
		user     *domain.User
		findErr  error
		email    string
		password string
		wantErr  error
		verify   func(t *testing.T, result *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher)
	}{
		{
			name:     "creates an admin when the email is missing",
			email:    "root@milpa.com",
			password: "supersecret",
			verify: func(t *testing.T, result *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()

				if result == nil {
					t.Fatal("result is nil")
				}
				if !result.Created {
					t.Error("Created = false, want true")
				}
				if result.Promoted {
					t.Error("Promoted = true, want false")
				}
				if result.Email != "root@milpa.com" {
					t.Errorf("Email = %q, want %q", result.Email, "root@milpa.com")
				}
				if result.UserID == uuid.Nil {
					t.Error("UserID is nil, want a generated uuid")
				}

				if repo.saveCalls != 1 {
					t.Errorf("Save calls = %d, want 1", repo.saveCalls)
				}
				if repo.updateCalls != 0 {
					t.Errorf("Update calls = %d, want 0", repo.updateCalls)
				}
				if hasher.hashCalls != 1 {
					t.Errorf("Hash calls = %d, want 1", hasher.hashCalls)
				}

				saved := repo.savedUser
				if saved == nil {
					t.Fatal("no user was saved")
				}
				if saved.Role != domain.RoleAdmin {
					t.Errorf("saved role = %v, want admin", saved.Role)
				}
				if saved.PasswordHash != "hashed:supersecret" {
					t.Errorf("saved password hash = %q, want %q", saved.PasswordHash, "hashed:supersecret")
				}
				if saved.FirstName != "Admin" || saved.LastName != "Milpa" {
					t.Errorf("saved name = %q %q, want Admin Milpa", saved.FirstName, saved.LastName)
				}
				if !saved.CreatedAt.Equal(fixedNow) || !saved.UpdatedAt.Equal(fixedNow) {
					t.Errorf("saved timestamps = %v / %v, want %v", saved.CreatedAt, saved.UpdatedAt, fixedNow)
				}
				if result.UserID != saved.ID {
					t.Errorf("result UserID = %v, want %v", result.UserID, saved.ID)
				}
			},
		},
		{
			name:     "does nothing when the user is already admin",
			user:     &domain.User{ID: adminID, Email: "root@milpa.com", Role: domain.RoleAdmin, PasswordHash: "existing-hash", UpdatedAt: existingTime},
			email:    "root@milpa.com",
			password: "supersecret",
			verify: func(t *testing.T, result *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()

				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Created || result.Promoted {
					t.Errorf("Created = %v, Promoted = %v, want both false", result.Created, result.Promoted)
				}
				if result.UserID != adminID {
					t.Errorf("UserID = %v, want %v", result.UserID, adminID)
				}
				if result.Email != "root@milpa.com" {
					t.Errorf("Email = %q, want %q", result.Email, "root@milpa.com")
				}
				if repo.updateCalls != 0 {
					t.Errorf("Update calls = %d, want 0", repo.updateCalls)
				}
				if repo.saveCalls != 0 {
					t.Errorf("Save calls = %d, want 0", repo.saveCalls)
				}
				if hasher.hashCalls != 0 {
					t.Errorf("Hash calls = %d, want 0", hasher.hashCalls)
				}
				if repo.user.PasswordHash != "existing-hash" {
					t.Errorf("password hash = %q, want it untouched", repo.user.PasswordHash)
				}
				if !repo.user.UpdatedAt.Equal(existingTime) {
					t.Errorf("UpdatedAt = %v, want it untouched (%v)", repo.user.UpdatedAt, existingTime)
				}
			},
		},
		{
			name:     "promotes an existing non-admin without touching the password",
			user:     &domain.User{ID: farmerID, Email: "farmer@milpa.com", Role: domain.RoleAgricultor, PasswordHash: "existing-hash", UpdatedAt: existingTime},
			email:    "farmer@milpa.com",
			password: "supersecret",
			verify: func(t *testing.T, result *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()

				if result == nil {
					t.Fatal("result is nil")
				}
				if !result.Promoted {
					t.Error("Promoted = false, want true")
				}
				if result.Created {
					t.Error("Created = true, want false")
				}
				if result.UserID != farmerID {
					t.Errorf("UserID = %v, want %v", result.UserID, farmerID)
				}
				if repo.updateCalls != 1 {
					t.Errorf("Update calls = %d, want 1", repo.updateCalls)
				}
				if repo.saveCalls != 0 {
					t.Errorf("Save calls = %d, want 0", repo.saveCalls)
				}
				if hasher.hashCalls != 0 {
					t.Errorf("Hash calls = %d, want 0", hasher.hashCalls)
				}
				if repo.user.Role != domain.RoleAdmin {
					t.Errorf("role = %v, want admin", repo.user.Role)
				}
				if repo.user.PasswordHash != "existing-hash" {
					t.Errorf("password hash = %q, want it untouched", repo.user.PasswordHash)
				}
				if !repo.user.UpdatedAt.Equal(fixedNow) {
					t.Errorf("UpdatedAt = %v, want %v", repo.user.UpdatedAt, fixedNow)
				}
			},
		},
		{
			name:     "rejects an empty email",
			email:    "",
			password: "supersecret",
			wantErr:  domain.ErrInvalidInput,
			verify: func(t *testing.T, _ *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()
				if repo.saveCalls != 0 || repo.updateCalls != 0 || hasher.hashCalls != 0 {
					t.Errorf("repo/hasher were used: save=%d update=%d hash=%d", repo.saveCalls, repo.updateCalls, hasher.hashCalls)
				}
			},
		},
		{
			name:     "rejects an empty password",
			email:    "root@milpa.com",
			password: "",
			wantErr:  domain.ErrInvalidInput,
			verify: func(t *testing.T, _ *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()
				if repo.saveCalls != 0 || repo.updateCalls != 0 || hasher.hashCalls != 0 {
					t.Errorf("repo/hasher were used: save=%d update=%d hash=%d", repo.saveCalls, repo.updateCalls, hasher.hashCalls)
				}
			},
		},
		{
			name:     "rejects a password shorter than eight characters",
			email:    "root@milpa.com",
			password: "short",
			wantErr:  domain.ErrInvalidInput,
			verify: func(t *testing.T, _ *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()
				if repo.saveCalls != 0 || repo.updateCalls != 0 || hasher.hashCalls != 0 {
					t.Errorf("repo/hasher were used: save=%d update=%d hash=%d", repo.saveCalls, repo.updateCalls, hasher.hashCalls)
				}
			},
		},
		{
			name:     "propagates a lookup failure",
			findErr:  findFailure,
			email:    "root@milpa.com",
			password: "supersecret",
			wantErr:  findFailure,
			verify: func(t *testing.T, _ *BootstrapResult, repo *fakeUserRepo, hasher *fakeHasher) {
				t.Helper()
				if repo.saveCalls != 0 || repo.updateCalls != 0 || hasher.hashCalls != 0 {
					t.Errorf("repo/hasher were used: save=%d update=%d hash=%d", repo.saveCalls, repo.updateCalls, hasher.hashCalls)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeUserRepo{user: tt.user, findErr: tt.findErr}
			hasher := &fakeHasher{}
			uc := NewBootstrapAdminUseCase(repo, hasher, fakeClock{now: fixedNow})

			result, err := uc.EnsureAdmin(context.Background(), tt.email, tt.password, "Admin", "Milpa")

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("EnsureAdmin() error = %v, want %v", err, tt.wantErr)
				}
				if result != nil {
					t.Fatalf("EnsureAdmin() result = %+v, want nil on error", result)
				}
			} else if err != nil {
				t.Fatalf("EnsureAdmin() unexpected error = %v", err)
			}

			tt.verify(t, result, repo, hasher)
		})
	}
}
