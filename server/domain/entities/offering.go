package domain

import (
	"fmt"
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
	UserID      uuid.UUID
	Type        OfferingType
	Name        string
	Description string
	Price       float64
	ImageURL    string

	Variety           string
	UnitOfMeasureID   *uuid.UUID
	QuantityAvailable float64
	ExpiresAt         *time.Time
	IsActive          bool
	CategoryID        *uuid.UUID
	CompanyID         *uuid.UUID
	Latitude          *float64
	Longitude         *float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Builder

func NewOffering(userID uuid.UUID, name string, offeringType OfferingType, now time.Time) (*Offering, error) {
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
		UserID:    userID,
		Type:      offeringType,
		Name:      name,
		IsActive:  true,
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

func (o Offering) IsExpired(now time.Time) bool {
	return o.ExpiresAt != nil && !o.ExpiresAt.After(now)
}

func (o Offering) IsVisibleAt(now time.Time) bool {
	return o.IsActive && !o.IsExpired(now)
}

func (o Offering) HasLocation() bool {
	return o.Latitude != nil && o.Longitude != nil
}

func (o Offering) RequirePublishable() error {
	if o.Variety == "" {
		return ErrVarietyRequired
	}
	if o.UnitOfMeasureID == nil {
		return ErrUnitOfMeasureRequired
	}
	if o.QuantityAvailable <= 0 {
		return ErrInvalidQuantity
	}
	if o.CategoryID == nil {
		return ErrCategoryRequired
	}
	return nil
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

func (o *Offering) SetVariety(variety string, now time.Time) {
	o.Variety = variety
	o.Touch(now)
}

func (o *Offering) SetUnitOfMeasure(unitID *uuid.UUID, now time.Time) {
	o.UnitOfMeasureID = unitID
	o.Touch(now)
}

func (o *Offering) SetQuantity(quantity float64, now time.Time) {
	o.QuantityAvailable = quantity
	o.Touch(now)
}

func (o *Offering) SetCategory(categoryID *uuid.UUID, now time.Time) {
	o.CategoryID = categoryID
	o.Touch(now)
}

func (o *Offering) SetCompany(companyID *uuid.UUID, now time.Time) {
	o.CompanyID = companyID
	o.Touch(now)
}

func (o *Offering) SetExpiry(expiresAt *time.Time, now time.Time) {
	o.ExpiresAt = expiresAt
	o.Touch(now)
}

func (o *Offering) SetLocation(latitude, longitude *float64, now time.Time) error {
	if err := validateLocation(latitude, longitude); err != nil {
		return err
	}
	o.Latitude = latitude
	o.Longitude = longitude
	o.Touch(now)
	return nil
}

func (o *Offering) Deactivate() {
	o.IsActive = false
}

func (o *Offering) Renew(expiresAt time.Time, now time.Time) {
	o.ExpiresAt = &expiresAt
	o.IsActive = true
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

func validateLocation(latitude, longitude *float64) error {
	if (latitude == nil) != (longitude == nil) {
		return fmt.Errorf("%w: latitude and longitude must be supplied together", ErrInvalidInput)
	}
	if latitude == nil {
		return nil
	}
	if *latitude < -90 || *latitude > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidInput)
	}
	if *longitude < -180 || *longitude > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidInput)
	}
	return nil
}
