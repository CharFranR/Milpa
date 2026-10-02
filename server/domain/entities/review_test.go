package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewReview(t *testing.T) {
	t.Parallel()

	now := time.Now()
	authorID := uuid.New()
	companyID := uuid.New()

	tests := []struct {
		name          string
		authorID      uuid.UUID
		targetType    ReviewTargetType
		targetID      uuid.UUID
		companyID     uuid.UUID
		rating        int
		comment       string
		transactionID uuid.UUID
		wantErr       error
	}{
		{name: "happy path lower boundary", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 1, comment: "Bad", transactionID: uuid.New()},
		{name: "happy path upper boundary", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 5, comment: "Great", transactionID: uuid.New()},
		{name: "rating below range", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 0, transactionID: uuid.New(), wantErr: ErrInvalidRating},
		{name: "rating above range", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 6, transactionID: uuid.New(), wantErr: ErrInvalidRating},
		{name: "no author", authorID: uuid.Nil, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 4, transactionID: uuid.New(), wantErr: ErrAuthorRequired},
		{name: "no transaction", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: companyID, rating: 4, wantErr: ErrTransactionRequired},
		{name: "unknown target type", authorID: authorID, targetType: "transaction", targetID: companyID, rating: 4, transactionID: uuid.New(), wantErr: ErrInvalidReviewTargetType},
		{name: "no target", authorID: authorID, targetType: ReviewTargetUser, rating: 4, transactionID: uuid.New(), wantErr: ErrTargetRequired},
		{name: "self rating", authorID: authorID, targetType: ReviewTargetUser, targetID: authorID, rating: 5, transactionID: uuid.New(), wantErr: ErrSelfReview},
		{name: "company mirror does not match target", authorID: authorID, targetType: ReviewTargetCompany, targetID: companyID, companyID: uuid.New(), rating: 4, transactionID: uuid.New(), wantErr: ErrReviewTargetMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			review, err := NewReview(tt.authorID, tt.targetType, tt.targetID, tt.companyID, tt.rating, tt.comment, now, tt.transactionID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if review.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if review.AuthorID != tt.authorID {
				t.Errorf("author id = %v, want %v", review.AuthorID, tt.authorID)
			}
			if review.TargetType != tt.targetType {
				t.Errorf("target type = %v, want %v", review.TargetType, tt.targetType)
			}
			if review.TargetID != tt.targetID {
				t.Errorf("target id = %v, want %v", review.TargetID, tt.targetID)
			}
			if review.CompanyID != tt.companyID {
				t.Errorf("company id = %v, want %v", review.CompanyID, tt.companyID)
			}
			if review.Rating != tt.rating {
				t.Errorf("rating = %v, want %v", review.Rating, tt.rating)
			}
			if review.Comment != tt.comment {
				t.Errorf("comment = %q, want %q", review.Comment, tt.comment)
			}
			if !review.CreatedAt.Equal(now) {
				t.Errorf("created at = %v, want %v", review.CreatedAt, now)
			}
		})
	}
}

// TestNewReviewUserTargetDropsTheCompanyMirror is the entity half of
// ck_reviews_company_mirror: a review of a farmer has no company, and a caller
// that passes one anyway must not have it silently persisted as that farmer's
// company rating.
func TestNewReviewUserTargetDropsTheCompanyMirror(t *testing.T) {
	t.Parallel()

	review, err := NewReview(uuid.New(), ReviewTargetUser, uuid.New(), uuid.New(), 4, "Reliable delivery", time.Now(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.CompanyID != uuid.Nil {
		t.Errorf("company id = %v, want the zero uuid for a user review", review.CompanyID)
	}
}

func TestValidReviewTargetType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		targetType ReviewTargetType
		want       bool
	}{
		{targetType: ReviewTargetCompany, want: true},
		{targetType: ReviewTargetUser, want: true},
		{targetType: "", want: false},
		{targetType: "Company", want: false},
		{targetType: "transaction", want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.targetType), func(t *testing.T) {
			t.Parallel()
			if got := ValidReviewTargetType(tt.targetType); got != tt.want {
				t.Errorf("ValidReviewTargetType(%q) = %v, want %v", tt.targetType, got, tt.want)
			}
		})
	}
}
