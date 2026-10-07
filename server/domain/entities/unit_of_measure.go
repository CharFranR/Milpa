package domain

import "github.com/google/uuid"

type UnitOfMeasure struct {
	ID       uuid.UUID
	Code     string
	Name     string
	IsActive bool
}