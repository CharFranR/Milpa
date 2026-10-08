package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

const minAdminPasswordLength = 8

// The user repository requires a non-empty phone number for every persisted
// user (Save validates it) and the bootstrap has no real phone to collect.
// A clearly fake placeholder keeps that contract; an operator can replace it
// later through the profile update endpoint.
const placeholderAdminPhone = "+505 0000 0000"

type BootstrapAdminUseCaseImpl struct {
	userRepo port.UserRepository
	hasher   port.PasswordHasher
	timer    port.TimeProvider
}

func NewBootstrapAdminUseCase(userRepo port.UserRepository, hasher port.PasswordHasher, timer port.TimeProvider) *BootstrapAdminUseCaseImpl {
	return &BootstrapAdminUseCaseImpl{
		userRepo: userRepo,
		hasher:   hasher,
		timer:    timer,
	}
}

type BootstrapResult struct {
	UserID   uuid.UUID
	Email    string
	Created  bool
	Promoted bool
}

// EnsureAdmin creates the administrator for the given email when it does not
// exist, promotes it when it exists with another role, and does nothing when it
// is already an admin. It never resets the password of an existing user.
func (uc *BootstrapAdminUseCaseImpl) EnsureAdmin(ctx context.Context, email, password, firstName, lastName string) (*BootstrapResult, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}
	if password == "" {
		return nil, fmt.Errorf("%w: password is required", domain.ErrInvalidInput)
	}
	if len(password) < minAdminPasswordLength {
		return nil, fmt.Errorf("%w: password must be at least %d characters", domain.ErrInvalidInput, minAdminPasswordLength)
	}

	now := uc.timer.Now()

	user, err := uc.userRepo.FindByEmail(ctx, email)
	switch {
	case err == nil:
		if user.Role == domain.RoleAdmin {
			return &BootstrapResult{UserID: user.ID, Email: user.Email}, nil
		}

		user.Role = domain.RoleAdmin
		user.Touch(now)

		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}

		return &BootstrapResult{UserID: user.ID, Email: user.Email, Promoted: true}, nil
	case errors.Is(err, domain.ErrNotFound):
		user, err := domain.NewUser(email, firstName, lastName, now)
		if err != nil {
			return nil, err
		}
		user.Role = domain.RoleAdmin

		hash, err := uc.hasher.Hash(password)
		if err != nil {
			return nil, err
		}
		user.SetPasswordHash(hash)
		user.PhoneNumber = placeholderAdminPhone

		if _, err := uc.userRepo.Save(ctx, user); err != nil {
			return nil, err
		}

		return &BootstrapResult{UserID: user.ID, Email: user.Email, Created: true}, nil
	default:
		return nil, err
	}
}
