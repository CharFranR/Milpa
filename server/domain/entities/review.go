package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReviewTargetType string

const (
	ReviewTargetCompany ReviewTargetType = "company"
	ReviewTargetUser    ReviewTargetType = "user"
)

func ValidReviewTargetType(t ReviewTargetType) bool {
	switch t {
	case ReviewTargetCompany, ReviewTargetUser:
		return true
	default:
		return false
	}
}

type Review struct {
	ID            uuid.UUID
	AuthorID      uuid.UUID
	TargetType    ReviewTargetType
	TargetID      uuid.UUID
	CompanyID     uuid.UUID
	Rating        int
	Comment       string
	CreatedAt     time.Time
	TransactionID uuid.UUID
}

func NewReview(
	authorID uuid.UUID, targetType ReviewTargetType, targetID, companyID uuid.UUID, rating int, comment string, now time.Time, transactionID uuid.UUID,
) (*Review, error) {
	if authorID == uuid.Nil {
		return nil, ErrAuthorRequired
	}
	if transactionID == uuid.Nil {
		return nil, ErrTransactionRequired
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
		ID:            uuid.New(),
		AuthorID:      authorID,
		TargetType:    targetType,
		TargetID:      targetID,
		CompanyID:     companyID,
		Rating:        rating,
		Comment:       comment,
		CreatedAt:     now,
		TransactionID: transactionID,
	}, nil
}
