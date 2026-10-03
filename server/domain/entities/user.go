package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoleOptions int

const (
	RolePending RoleOptions = iota
	RoleAgricultor
	RoleCompradorMinorista
	RoleCompradorMayoristaDetallista
	RoleCompradorMayoristaCorporativo
	RoleAdmin
	RoleAuditor
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
	SuspendedAt  *time.Time
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
	case RoleAgricultor:
		return "agricultor"
	case RoleCompradorMinorista:
		return "comprador_minorista"
	case RoleCompradorMayoristaDetallista:
		return "comprador_mayorista_detallista"
	case RoleCompradorMayoristaCorporativo:
		return "comprador_mayorista_corporativo"
	case RoleAdmin:
		return "admin"
	case RoleAuditor:
		return "auditor"
	default:
		return "unknown"
	}
}

func IsRegistrationRole(r RoleOptions) bool {
	switch r {
	case RoleAgricultor, RoleCompradorMinorista, RoleCompradorMayoristaDetallista, RoleCompradorMayoristaCorporativo:
		return true
	default:
		return false
	}
}

func ValidRole(r RoleOptions) bool {
	switch r {
	case RolePending, RoleAgricultor, RoleCompradorMinorista, RoleCompradorMayoristaDetallista, RoleCompradorMayoristaCorporativo, RoleAdmin, RoleAuditor:
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

func (u *User) Suspend(now time.Time) {
	u.SuspendedAt = &now
	u.Touch(now)
}

func (u *User) Reactivate(now time.Time) {
	u.SuspendedAt = nil
	u.Touch(now)
}

func (u User) IsSuspended() bool {
	return u.SuspendedAt != nil
}
