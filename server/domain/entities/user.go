package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoleOptions int

const (
	RolePending RoleOptions = iota
	RoleMIPYME
	RoleProvider
	RoleAdmin
)

type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Role      RoleOptions
	CreatedAt time.Time
	UpdatedAt time.Time
	Address   Address

	Email        string
	PhoneNumber  string
	PasswordHash string
}

// Builder

func NewUser(email, firstName, lastName string, now time.Time) (*User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}
	if firstName == "" {
		return nil, ErrFirstNameRequired
	}
	if lastName == "" {
		return nil, ErrLastNameRequired
	}

	return &User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      RolePending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get

func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u User) HasRole(role RoleOptions) bool {
	return u.Role == role
}

func (r RoleOptions) String() string {
	switch r {
	case RolePending:
		return "pending"
	case RoleMIPYME:
		return "mipyme"
	case RoleProvider:
		return "provider"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func ValidRole(r RoleOptions) bool {
	switch r {
	case RolePending, RoleMIPYME, RoleProvider, RoleAdmin:
		return true
	default:
		return false
	}
}

// Set

func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
}

func (u *User) Touch(now time.Time) {
	u.UpdatedAt = now
}
