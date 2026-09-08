package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

type CachedUserUseCase struct {
	next  primary.UserUseCase
	cache port.Cache
}

func NewCachedUserUseCase(next primary.UserUseCase, cache port.Cache) *CachedUserUseCase {
	return &CachedUserUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedUserUseCase) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.UserDTO, error) {
	return uc.next.Register(ctx, req)
}

func (uc *CachedUserUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return uc.next.Login(ctx, req)
}

func (uc *CachedUserUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserDTO, error) {
	var user *dto.UserDTO

	_, err := uc.cache.Remember(
		ctx,
		"user:"+id.String(),
		5*time.Minute,
		&user,
		func() error {
			result, err := uc.next.GetByID(ctx, id)
			if err != nil {
				return err
			}

			user = result
			return nil
		},
	)

	return user, err
}

func (uc *CachedUserUseCase) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	err := uc.next.UpdateProfile(ctx, id, req)
	if err != nil {
		return err
	}

	_ = uc.cache.Delete(ctx, "user:"+id.String())

	return nil
}

var _ primary.UserUseCase = (*CachedUserUseCase)(nil)
