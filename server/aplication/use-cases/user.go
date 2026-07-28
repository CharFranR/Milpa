package usecases

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
)

type UserUseCaseImpl struct {
	userRepo port.UserRepository
	hasher   port.PasswordHasher
	jwt      port.JWTProvider
	timer    port.TimeProvider
}

func NewUserUseCase(
	userRepo port.UserRepository,
	hasher port.PasswordHasher,
	jwt port.JWTProvider,
	timer port.TimeProvider,
) *UserUseCaseImpl {
	return &UserUseCaseImpl{
		userRepo: userRepo,
		hasher:   hasher,
		jwt:      jwt,
		timer:    timer,
	}
}

func (uc *UserUseCaseImpl) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.UserDTO, error) {
	now := uc.timer.Now()

	if req.Role != domain.RoleMIPYME && req.Role != domain.RoleProvider {
		return nil, domain.ErrInvalidInput
	}

	if req.Password != req.ConfirmPassword {
		return nil, domain.ErrInvalidInput
	}

	user, err := domain.NewUser(req.Email, req.FirstName, req.LastName, now)
	if err != nil {
		return nil, err
	}

	user.Role = req.Role

	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	hash, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	user.SetPasswordHash(hash)

	user.PhoneNumber = req.PhoneNumber

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return userToDTO(user), nil
}

func (uc *UserUseCaseImpl) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := uc.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrUnauthorized
	}

	token, err := uc.jwt.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: token,
		ExpiresIn:   86400,
		User:        *userToDTO(user),
	}, nil
}

func (uc *UserUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserDTO, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return userToDTO(user), nil
}

func (uc *UserUseCaseImpl) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Address != nil {
		user.Address = domain.Address{AddressLine: *req.Address}
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}

	user.Touch(uc.timer.Now())

	return uc.userRepo.Update(ctx, user)
}

var _ primary.UserUseCase = (*UserUseCaseImpl)(nil)

func userToDTO(user *domain.User) *dto.UserDTO {
	return &dto.UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Address:     user.Address.FullAddress(),
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
