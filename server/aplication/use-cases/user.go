package usecases

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
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

func (uc *UserUseCaseImpl) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.PrivateUserDTO, error) {
	now := uc.timer.Now()

	if !domain.IsRegistrationRole(req.Role) {
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
	user.Address = domain.Address{
		AddressLine:  req.Address,
		Department:   req.Department,
		Municipality: req.Municipality,
		Latitude:     floatOrZero(req.Latitude),
		Longitude:    floatOrZero(req.Longitude),
	}
	if err := user.Address.ValidateCoordinates(); err != nil {
		return nil, err
	}

	if _, err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// The caller is the user that was just created, so this is the private view.
	return privateUserDTO(user), nil
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
		User:        *privateUserDTO(user),
	}, nil
}

func (uc *UserUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.UserView, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return userViewFor(ctx, user), nil
}

func (uc *UserUseCaseImpl) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if user.ID != principal.UserID {
		return domain.ErrForbidden
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
		user.Address.AddressLine = *req.Address
	}
	if req.Department != nil {
		user.Address.Department = *req.Department
	}
	if req.Municipality != nil {
		user.Address.Municipality = *req.Municipality
	}
	if req.Latitude != nil {
		user.Address.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		user.Address.Longitude = *req.Longitude
	}
	if err := user.Address.ValidateCoordinates(); err != nil {
		return err
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}

	user.Touch(uc.timer.Now())

	return uc.userRepo.Update(ctx, user)
}

var _ primary.UserUseCase = (*UserUseCaseImpl)(nil)

func floatOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func userViewFor(ctx context.Context, user *domain.User) dto.UserView {
	principal, ok := auth.FromContext(ctx)
	if ok && (principal.UserID == user.ID || principal.Role == domain.RoleAdmin) {
		return privateUserDTO(user)
	}
	return publicUserDTO(user)
}

func publicUserDTO(user *domain.User) *dto.PublicUserDTO {
	return &dto.PublicUserDTO{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         user.Role,
		Department:   user.Address.Department,
		Municipality: user.Address.Municipality,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func privateUserDTO(user *domain.User) *dto.PrivateUserDTO {
	return &dto.PrivateUserDTO{
		PublicUserDTO: *publicUserDTO(user),
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		AddressLine:   user.Address.AddressLine,
	}
}
