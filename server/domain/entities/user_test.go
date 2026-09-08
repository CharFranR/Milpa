package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		email     string
		firstName string
		lastName  string
		now       time.Time
		wantErr   error
	}{
		{name: "happy path", email: "example@milpa.com.ni", firstName: "John", lastName: "Doe", now: now},
		{name: "empty email", firstName: "John", lastName: "Doe", now: now, wantErr: ErrEmailRequired},
		{name: "empty first name", email: "example@milpa.com.ni", lastName: "Doe", now: now, wantErr: ErrFirstNameRequired},
		{name: "empty last name", email: "example@milpa.com.ni", firstName: "John", now: now, wantErr: ErrLastNameRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := NewUser(tt.email, tt.firstName, tt.lastName, tt.now)

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
			if user.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if user.FirstName != tt.firstName {
				t.Errorf("first name = %q, want %q", user.FirstName, tt.firstName)
			}
			if user.LastName != tt.lastName {
				t.Errorf("last name = %q, want %q", user.LastName, tt.lastName)
			}
			if user.Email != tt.email {
				t.Errorf("email = %q, want %q", user.Email, tt.email)
			}
			if !user.CreatedAt.Equal(tt.now) {
				t.Errorf("created at = %v, want %v", user.CreatedAt, tt.now)
			}
			if !user.UpdatedAt.Equal(tt.now) {
				t.Errorf("updated at = %v, want %v", user.UpdatedAt, tt.now)
			}
			if user.Role != RolePending {
				t.Errorf("role = %v, want %v", user.Role, RolePending)
			}
			if user.Address != (Address{}) {
				t.Errorf("address = %+v, want zero value", user.Address)
			}
			if user.PhoneNumber != "" {
				t.Errorf("phone number = %q, want empty", user.PhoneNumber)
			}
			if user.PasswordHash != "" {
				t.Errorf("password hash = %q, want empty", user.PasswordHash)
			}
		})
	}
}

func TestUserFullName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{name: "both names", firstName: "John", lastName: "Doe", want: "John Doe"},
		{name: "first name only", firstName: "Jane", want: "Jane "},
		{name: "last name only", lastName: "Doe", want: " Doe"},
		{name: "zero value", want: " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := User{FirstName: tt.firstName, LastName: tt.lastName}

			if got := user.FullName(); got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserIsAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		role RoleOptions
		want bool
	}{
		{name: "admin role", role: RoleAdmin, want: true},
		{name: "pending role", role: RolePending, want: false},
		{name: "mipyme role", role: RoleMIPYME, want: false},
		{name: "provider role", role: RoleProvider, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := User{Role: tt.role}

			if got := user.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserHasRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     RoleOptions
		expected RoleOptions
		want     bool
	}{
		{name: "matching role", role: RoleProvider, expected: RoleProvider, want: true},
		{name: "different role", role: RolePending, expected: RoleAdmin, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := User{Role: tt.role}

			if got := user.HasRole(tt.expected); got != tt.want {
				t.Errorf("HasRole(%v) = %v, want %v", tt.expected, got, tt.want)
			}
		})
	}
}

func TestRoleOptionsString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		role RoleOptions
		want string
	}{
		{name: "pending", role: RolePending, want: "pending"},
		{name: "mipyme", role: RoleMIPYME, want: "mipyme"},
		{name: "provider", role: RoleProvider, want: "provider"},
		{name: "admin", role: RoleAdmin, want: "admin"},
		{name: "unknown value", role: RoleOptions(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.role.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserSetPasswordHash(t *testing.T) {
	t.Parallel()

	want := "hashed-password"
	user := &User{}

	user.SetPasswordHash(want)

	if user.PasswordHash != want {
		t.Errorf("PasswordHash = %q, want %q", user.PasswordHash, want)
	}
}

func TestUserTouch(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name string
		user User
	}{
		{name: "zero value", user: User{}},
		{name: "updates previous timestamp", user: User{UpdatedAt: earlier}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := tt.user
			user.Touch(now)

			if !user.UpdatedAt.Equal(now) {
				t.Errorf("UpdatedAt = %v, want %v", user.UpdatedAt, now)
			}
		})
	}
}
