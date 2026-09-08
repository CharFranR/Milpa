package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewCompany(t *testing.T) {
	t.Parallel()

	now := time.Now()
	owner := User{ID: uuid.New()}

	tests := []struct {
		name    string
		owner   User
		company string
		now     time.Time
		wantErr error
	}{
		{name: "happy path", owner: owner, company: "Milpa Foods", now: now},
		{name: "owner with nil UUID", company: "Milpa Foods", now: now, wantErr: ErrOwnerRequired},
		{name: "empty name", owner: owner, now: now, wantErr: ErrNameRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			company, err := NewCompany(tt.owner, tt.company, tt.now)

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
			if company.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if company.Name != tt.company {
				t.Errorf("name = %q, want %q", company.Name, tt.company)
			}
			if company.Owner != tt.owner {
				t.Errorf("owner = %+v, want %+v", company.Owner, tt.owner)
			}
			if company.Verified {
				t.Error("verified = true, want false")
			}
			if !company.CreatedAt.Equal(tt.now) {
				t.Errorf("created at = %v, want %v", company.CreatedAt, tt.now)
			}
			if !company.UpdatedAt.Equal(tt.now) {
				t.Errorf("updated at = %v, want %v", company.UpdatedAt, tt.now)
			}
			if company.Category != nil {
				t.Errorf("category = %+v, want nil", company.Category)
			}
			if company.Address != (Address{}) {
				t.Errorf("address = %+v, want zero value", company.Address)
			}
			if company.Description != "" {
				t.Errorf("description = %q, want empty", company.Description)
			}
			if company.PhoneNumber != "" {
				t.Errorf("phone number = %q, want empty", company.PhoneNumber)
			}
			if company.Email != "" {
				t.Errorf("email = %q, want empty", company.Email)
			}
			if company.Website != "" {
				t.Errorf("website = %q, want empty", company.Website)
			}
		})
	}
}

func TestCompanyIsVerified(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		verify bool
		want   bool
	}{
		{name: "not verified by default", want: false},
		{name: "verified after Verify", verify: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			company := &Company{}
			if tt.verify {
				company.Verify()
			}

			if got := company.IsVerified(); got != tt.want {
				t.Errorf("IsVerified() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanyVerify(t *testing.T) {
	t.Parallel()

	company := &Company{}

	company.Verify()

	if !company.Verified {
		t.Error("Verified = false, want true")
	}
}

func TestCompanyTouch(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name    string
		company Company
	}{
		{name: "zero value", company: Company{}},
		{name: "updates previous timestamp", company: Company{UpdatedAt: earlier}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			company := tt.company
			company.Touch(now)

			if !company.UpdatedAt.Equal(now) {
				t.Errorf("UpdatedAt = %v, want %v", company.UpdatedAt, now)
			}
		})
	}
}

func TestCompanyChangeOwner(t *testing.T) {
	t.Parallel()

	currentOwner := User{ID: uuid.New()}
	newOwner := User{ID: uuid.New()}

	tests := []struct {
		name     string
		newOwner User
		wantErr  error
	}{
		{name: "valid new owner", newOwner: newOwner},
		{name: "new owner with nil UUID", newOwner: User{}, wantErr: ErrOwnerRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			company := &Company{Owner: currentOwner}

			err := company.ChangeOwner(tt.newOwner)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if company.Owner != currentOwner {
					t.Errorf("owner = %+v, want %+v", company.Owner, currentOwner)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if company.Owner != tt.newOwner {
				t.Errorf("owner = %+v, want %+v", company.Owner, tt.newOwner)
			}
		})
	}
}
