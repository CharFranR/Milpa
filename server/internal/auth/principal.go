package auth

import (
	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type Principal struct {
	UserID uuid.UUID
	Role   domain.RoleOptions
}
