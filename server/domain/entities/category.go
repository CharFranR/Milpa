package domain

import "github.com/google/uuid"

const DefaultMainCategory = "otros"

type Category struct {
	ID                     uuid.UUID
	Name                   string
	Description            string
	MainCategory           string
	IsActive               bool
	DefaultUnitOfMeasureID *uuid.UUID
}

func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	return &Category{
		ID:           uuid.New(),
		Name:         name,
		MainCategory: DefaultMainCategory,
		IsActive:     true,
	}, nil
}

func (c *Category) Deactivate() {
	c.IsActive = false
}

func (c *Category) Activate() {
	c.IsActive = true
}
