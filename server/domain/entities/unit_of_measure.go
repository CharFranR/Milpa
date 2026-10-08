package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type UnitOfMeasure struct {
	ID       uuid.UUID
	Code     string
	Name     string
	IsActive bool
}

func NewUnitOfMeasure(code, name string) (*UnitOfMeasure, error) {
	unit := &UnitOfMeasure{
		ID:       uuid.New(),
		IsActive: true,
	}

	if err := unit.SetCode(code); err != nil {
		return nil, err
	}
	if err := unit.SetName(name); err != nil {
		return nil, err
	}

	return unit, nil
}

func (u *UnitOfMeasure) SetCode(code string) error {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if normalized == "" {
		return ErrCodeRequired
	}
	if len(normalized) > 20 {
		return fmt.Errorf("%w: code must be at most 20 characters", ErrInvalidInput)
	}

	u.Code = normalized
	return nil
}

func (u *UnitOfMeasure) SetName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrNameRequired
	}
	if len(trimmed) > 60 {
		return fmt.Errorf("%w: name must be at most 60 characters", ErrInvalidInput)
	}

	u.Name = trimmed
	return nil
}

func (u *UnitOfMeasure) Activate() {
	u.IsActive = true
}

func (u *UnitOfMeasure) Deactivate() {
	u.IsActive = false
}
