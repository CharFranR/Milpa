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
	userID := uuid.New()
	companyID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		companyID uuid.UUID
		rating    int
		comment   string
		now       time.Time
		wantErr   error
	}{
		{name: "happy path lower boundary", userID: userID, companyID: companyID, rating: 1, comment: "Bad", now: now},
		{name: "happy path upper boundary", userID: userID, companyID: companyID, rating: 5, comment: "Great", now: now},
		{name: "rating below range", userID: userID, companyID: companyID, rating: 0, comment: "Bad", now: now, wantErr: ErrInvalidRating},
		{name: "rating above range", userID: userID, companyID: companyID, rating: 6, comment: "Great", now: now, wantErr: ErrInvalidRating},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			review, err := NewReview(tt.userID, tt.companyID, tt.rating, tt.comment, tt.now)

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
			if review.UserID != tt.userID {
				t.Errorf("user id = %v, want %v", review.UserID, tt.userID)
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
			if !review.CreatedAt.Equal(tt.now) {
				t.Errorf("created at = %v, want %v", review.CreatedAt, tt.now)
			}
		})
	}
}
