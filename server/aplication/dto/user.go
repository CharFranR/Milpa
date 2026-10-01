package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type UserView interface {
	userView()
}

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

type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	ExpiresIn   int64          `json:"expires_in"`
	User        PrivateUserDTO `json:"user"`
}
