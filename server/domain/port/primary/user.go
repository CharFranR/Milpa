package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type UserUseCase interface {
	Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.PrivateUserDTO, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	// GetByID returns dto.PublicUserDTO or dto.PrivateUserDTO depending on
	// whether the caller is the owner or an admin.
	GetByID(ctx context.Context, id uuid.UUID) (dto.UserView, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error
}
