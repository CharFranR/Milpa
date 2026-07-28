package dto

import (
	"time"

	domain "github.com/CharFranR/Hackaton2026/domain/entities"
	"github.com/google/uuid"
)

type OfferingDTO struct {
	ID          uuid.UUID          `json:"id"`
	CompanyID   uuid.UUID          `json:"company_id"`
	Type        domain.OfferingType `json:"type"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Price       float64            `json:"price"`
	ImageURL    string             `json:"image_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOfferingRequest struct {
	CompanyID   uuid.UUID           `json:"company_id"`
	Type        domain.OfferingType `json:"type"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Price       float64             `json:"price"`
	ImageURL    string              `json:"image_url,omitempty"`
}

type UpdateOfferingRequest struct {
	Type        *domain.OfferingType `json:"type,omitempty"`
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Price       *float64             `json:"price,omitempty"`
	ImageURL    *string              `json:"image_url,omitempty"`
}
