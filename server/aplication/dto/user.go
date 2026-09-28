package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

// UserView is what reading a user returns.
//
// The concrete type is the privacy boundary, not a field that is sometimes
// present. A public marketplace page and a private contact card are different
// things, and returning "the public one plus the private fields when allowed"
// as one struct with optional fields is how a field that leaks once ends up
// leaking forever. So the public type has no field for an email at all, which
// makes "an anonymous caller received an email" unrepresentable rather than
// merely untested.
type UserView interface {
	userView()
}

// PublicUserDTO is the marketplace representation of a user: identity, name,
// where they are, and what kind of account they hold. It carries no contact
// detail — reaching a producer goes through the platform, not around it.
type PublicUserDTO struct {
	ID           uuid.UUID          `json:"id"`
	FirstName    string             `json:"first_name"`
	LastName     string             `json:"last_name"`
	Role         domain.RoleOptions `json:"role"`
	Department   string             `json:"department"`
	Municipality string             `json:"municipality"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// PrivateUserDTO adds the contact card. It embeds the public representation so
// the shared shape is defined once and the two can never disagree about it.
type PrivateUserDTO struct {
	PublicUserDTO
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	AddressLine string `json:"address_line"`
}

func (PublicUserDTO) userView()  {}
func (PrivateUserDTO) userView() {}

type RegisterUserRequest struct {
	Email           string             `json:"email"`
	FirstName       string             `json:"first_name"`
	LastName        string             `json:"last_name"`
	Role            domain.RoleOptions `json:"role"`
	Address         string             `json:"address,omitempty"`
	Department      string             `json:"department,omitempty"`
	Municipality    string             `json:"municipality,omitempty"`
	Latitude        *float64           `json:"latitude,omitempty"`
	Longitude       *float64           `json:"longitude,omitempty"`
	PhoneNumber     string             `json:"phone_number,omitempty"`
	Password        string             `json:"password"`
	ConfirmPassword string             `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Email        *string  `json:"email,omitempty"`
	FirstName    *string  `json:"first_name,omitempty"`
	LastName     *string  `json:"last_name,omitempty"`
	Address      *string  `json:"address,omitempty"`
	Department   *string  `json:"department,omitempty"`
	Municipality *string  `json:"municipality,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	PhoneNumber  *string  `json:"phone_number,omitempty"`
}

// LoginResponse carries the caller's own profile, so it is the private view by
// construction: a caller can only get here by having authenticated as that
// user.
type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	ExpiresIn   int64          `json:"expires_in"`
	User        PrivateUserDTO `json:"user"`
}
