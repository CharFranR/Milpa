package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReviewDTO struct {
	ID uuid.UUID `json:"id"`
	// UserID is the review's author. The domain calls it AuthorID and the
	// column is author_id, but this was published as user_id before RF-15 and
	// renaming it here would break every consumer that reads it.
	UserID    uuid.UUID `json:"user_id"`
	CompanyID uuid.UUID `json:"company_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`

	// Additive since RF-15: without these a review of a user is
	// indistinguishable from a review of a company, CompanyID being zero.
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
}

type CreateReviewRequest struct {
	// CompanyID is how a client reviews a company. TargetType and TargetID are
	// how it reviews anyone else, and exactly one of the two forms is accepted,
	// so the two can never disagree about what is being reviewed.
	CompanyID  uuid.UUID `json:"company_id"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
}

// ReviewAverageDTO is the average rating of a target and how many reviews it is
// an average of. The count travels with it on purpose: 5.0 from two reviews and
// 5.0 from two hundred are not the same claim.
type ReviewAverageDTO struct {
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Average    float64   `json:"average"`
	Count      int       `json:"count"`
}
