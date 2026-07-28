package auth

import (
	domain "github.com/CharFranR/Hackaton2026/domain/entities"
	"github.com/google/uuid"
)

type Principal struct {
	UserID uuid.UUID
	Role   domain.RoleOptions
}
