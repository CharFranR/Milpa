package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReviewDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CompanyID uuid.UUID `json:"company_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	CompanyID uuid.UUID `json:"company_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
}
