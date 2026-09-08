package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CompanyID uuid.UUID
	Rating    int
	Comment   string

	CreatedAt time.Time
}

// Builder

func NewReview(userID, companyID uuid.UUID, rating int, comment string, now time.Time) (*Review, error) {
	if rating < 1 || rating > 5 {
		return nil, ErrInvalidRating
	}

	return &Review{
		ID:        uuid.New(),
		UserID:    userID,
		CompanyID: companyID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: now,
	}, nil
}
