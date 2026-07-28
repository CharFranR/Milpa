package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type UserUseCase interface {
	Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.UserDTO, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserDTO, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error
}
