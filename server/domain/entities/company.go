package domain

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID          uuid.UUID
	Name        string
	Category    []Category
	Owner       User
	Address     Address
	Description string
	PhoneNumber string
	Email       string
	Website     string
	Verified    bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Builder

func NewCompany(owner User, name string, now time.Time) (*Company, error) {
	if owner.ID == uuid.Nil {
		return nil, ErrOwnerRequired
	}
	if name == "" {
		return nil, ErrNameRequired
	}

	return &Company{
		ID:        uuid.New(),
		Owner:     owner,
		Name:      name,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get

func (c Company) IsVerified() bool {
	return c.Verified
}

// Set

func (c *Company) Verify() {
	c.Verified = true
}

func (c *Company) Touch(now time.Time) {
	c.UpdatedAt = now
}

func (c *Company) ChangeOwner(newOwner User) error {
	if newOwner.ID == uuid.Nil {
		return ErrOwnerRequired
	}

	c.Owner = newOwner
	return nil
}
