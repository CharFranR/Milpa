package domain

import "github.com/google/uuid"

type Category struct {
	ID          uuid.UUID
	Name        string
	Description string
}

func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	return &Category{
		ID:   uuid.New(),
		Name: name,
	}, nil
}
