package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type InquiryDTO struct {
	ID         uuid.UUID            `json:"id"`
	UserID     uuid.UUID            `json:"user_id"`
	OfferingID uuid.UUID            `json:"offering_id"`
	Message    string               `json:"message"`
	Status     domain.InquiryStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInquiryRequest struct {
	OfferingID uuid.UUID `json:"offering_id"`
	Message    string    `json:"message"`
}

type UpdateInquiryRequest struct {
	Status *domain.InquiryStatus `json:"status"`
}
