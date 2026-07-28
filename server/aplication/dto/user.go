package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID          uuid.UUID          `json:"id"`
	Email       string             `json:"email"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	Address     string             `json:"address"`
	PhoneNumber string             `json:"phone_number"`
	Role        domain.RoleOptions `json:"role"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterUserRequest struct {
	Email           string             `json:"email"`
	FirstName       string             `json:"first_name"`
	LastName        string             `json:"last_name"`
	Role            domain.RoleOptions `json:"role"`
	Address         string             `json:"address,omitempty"`
	PhoneNumber     string             `json:"phone_number,omitempty"`
	Password        string             `json:"password"`
	ConfirmPassword string             `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Email       *string `json:"email,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	Address     *string `json:"address,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type LoginResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   int64   `json:"expires_in"`
	User        UserDTO `json:"user"`
}
