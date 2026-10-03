package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type OfferingDTO struct {
	ID          uuid.UUID           `json:"id"`
	UserID      uuid.UUID           `json:"user_id"`
	Type        domain.OfferingType `json:"type"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       float64             `json:"price"`
	ImageURL    string              `json:"image_url"`

	Variety           string     `json:"variety"`
	UnitOfMeasureID   *uuid.UUID `json:"unit_of_measure_id,omitempty"`
	QuantityAvailable float64    `json:"quantity_available"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	IsActive          bool       `json:"is_active"`
	CategoryID        *uuid.UUID `json:"category_id,omitempty"`
	CompanyID         *uuid.UUID `json:"company_id,omitempty"`
	Latitude          *float64   `json:"latitude,omitempty"`
	Longitude         *float64   `json:"longitude,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOfferingRequest struct {
	UserID      uuid.UUID           `json:"user_id"`
	Type        domain.OfferingType `json:"type"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Price       float64             `json:"price"`
	ImageURL    string              `json:"image_url,omitempty"`

	Variety           string     `json:"variety"`
	UnitOfMeasureID   *uuid.UUID `json:"unit_of_measure_id"`
	QuantityAvailable float64    `json:"quantity_available"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CategoryID        *uuid.UUID `json:"category_id"`
	CompanyID         *uuid.UUID `json:"company_id,omitempty"`
	Latitude          *float64   `json:"latitude,omitempty"`
	Longitude         *float64   `json:"longitude,omitempty"`
}

type UpdateOfferingRequest struct {
	Type        *domain.OfferingType `json:"type,omitempty"`
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Price       *float64             `json:"price,omitempty"`
	ImageURL    *string              `json:"image_url,omitempty"`

	Variety           *string    `json:"variety,omitempty"`
	UnitOfMeasureID   *uuid.UUID `json:"unit_of_measure_id,omitempty"`
	QuantityAvailable *float64   `json:"quantity_available,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CategoryID        *uuid.UUID `json:"category_id,omitempty"`
	CompanyID         *uuid.UUID `json:"company_id,omitempty"`
	Latitude          *float64   `json:"latitude,omitempty"`
	Longitude         *float64   `json:"longitude,omitempty"`
}

type RenewOfferingRequest struct {
	ExpiresAt time.Time `json:"expires_at"`
}
