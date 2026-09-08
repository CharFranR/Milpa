package domain

import (
	"time"

	"github.com/google/uuid"
)

type OfferingType int

const (
	OfferingProduct OfferingType = iota
	OfferingService
)

type Offering struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Type        OfferingType
	Name        string
	Description string
	Price       float64
	ImageURL    string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Builder

func NewOffering(companyID uuid.UUID, name string, offeringType OfferingType, now time.Time) (*Offering, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	switch offeringType {
	case OfferingProduct, OfferingService:
	default:
		return nil, ErrInvalidOfferingType
	}

	return &Offering{
		ID:        uuid.New(),
		CompanyID: companyID,
		Type:      offeringType,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get

func (o Offering) IsProduct() bool {
	return o.Type == OfferingProduct
}

func (o Offering) IsService() bool {
	return o.Type == OfferingService
}

// Set

func (o *Offering) UpdatePrice(price float64, now time.Time) error {
	if price <= 0 {
		return ErrInvalidPrice
	}
	o.Price = price
	o.Touch(now)
	return nil
}

func (o *Offering) UpdateDescription(description string, now time.Time) {
	o.Description = description
	o.Touch(now)
}

func (o *Offering) UpdateImage(imageURL string, now time.Time) {
	o.ImageURL = imageURL
	o.Touch(now)
}

func (o *Offering) Touch(now time.Time) {
	o.UpdatedAt = now
}

func (ot OfferingType) String() string {
	switch ot {
	case OfferingProduct:
		return "product"
	case OfferingService:
		return "service"
	default:
		return "unknown"
	}
}
