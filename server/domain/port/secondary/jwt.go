package port

import (
	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type JWTProvider interface {
	GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

type JWTClaims struct {
	UserID uuid.UUID
	Role   domain.RoleOptions
}
