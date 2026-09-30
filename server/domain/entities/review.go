package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewTargetType names what a review is about. The vocabulary lives with the
// type so it has exactly one definition, the same way ReportTargetType does in
// report.go.
type ReviewTargetType string

const (
	ReviewTargetCompany ReviewTargetType = "company"
	ReviewTargetUser    ReviewTargetType = "user"
)

// ValidReviewTargetType reports whether t is part of the vocabulary, the same
// reason ValidMeasurementOptions exists: the field arrives off the wire as a
// string, so any string is representable, including ones nobody defined.
func ValidReviewTargetType(t ReviewTargetType) bool {
	switch t {
	case ReviewTargetCompany, ReviewTargetUser:
		return true
	default:
		return false
	}
}

type Review struct {
	ID       uuid.UUID
	AuthorID uuid.UUID

	TargetType ReviewTargetType
	TargetID   uuid.UUID

	// CompanyID mirrors TargetID for a company review and is uuid.Nil for a
	// user review, which has no company to mirror. The column is nullable for
	// the same reason and ck_reviews_company_mirror is what stops the two from
	// drifting apart in the database.
	CompanyID uuid.UUID

	Rating    int
	Comment   string
	CreatedAt time.Time
}

// NewReview follows NewReport's pattern: it validates everything that can be
// wrong about a review and returns the entity, rather than leaving the checks
// scattered across callers that each remember a different subset.
//
// companyID is the company mirror and must be uuid.Nil for a user review; for a
// company review it must equal targetID, which is the same invariant
// ck_reviews_company_mirror holds in SQL.
func NewReview(
	authorID uuid.UUID, targetType ReviewTargetType, targetID, companyID uuid.UUID, rating int, comment string, now time.Time,
) (*Review, error) {
	if authorID == uuid.Nil {
		return nil, ErrAuthorRequired
	}
	if targetType != ReviewTargetCompany && targetType != ReviewTargetUser {
		return nil, ErrInvalidReviewTargetType
	}
	if targetID == uuid.Nil {
		return nil, ErrTargetRequired
	}
	if rating < 1 || rating > 5 {
		return nil, ErrInvalidRating
	}
	// The same guard NewReport carries, for the same reason: a rating you gave
	// yourself is not a rating, and it would be the highest one on the profile
	// by construction.
	if authorID == targetID && targetType == ReviewTargetUser {
		return nil, ErrSelfReview
	}
	if targetType == ReviewTargetCompany && companyID != targetID {
		return nil, ErrReviewTargetMismatch
	}
	if targetType == ReviewTargetUser {
		companyID = uuid.Nil
	}

	return &Review{
		ID:         uuid.New(),
		AuthorID:   authorID,
		TargetType: targetType,
		TargetID:   targetID,
		CompanyID:  companyID,
		Rating:     rating,
		Comment:    comment,
		CreatedAt:  now,
	}, nil
}
