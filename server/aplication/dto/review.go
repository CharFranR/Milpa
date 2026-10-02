package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReviewDTO struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	CompanyID     uuid.UUID `json:"company_id"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
}

type CreateReviewRequest struct {
	CompanyID     uuid.UUID `json:"company_id"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	TransactionID uuid.UUID `json:"transaction_id"`
}

type ReviewAverageDTO struct {
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Average    float64   `json:"average"`
	Count      int       `json:"count"`
}
