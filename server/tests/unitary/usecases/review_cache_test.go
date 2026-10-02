package usecases_test

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

// A cached average that nothing evicts is a stale number on a profile forever,
// and it is the number a buyer trusts most.
func TestReviewCacheInvalidatesTheAggregate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		req          dto.CreateReviewRequest
		wantKeys     []string
		unwantedKeys []string
	}{
		{
			name: "a company review evicts the company's list and average",
			req:  dto.CreateReviewRequest{CompanyID: testCompanyID, Rating: 4, TransactionID: txTestTransactID},
			wantKeys: []string{
				"reviews:byuser:" + testUserID.String(),
				"reviews:bycompany:" + testCompanyID.String(),
				"reviews:avg:company:" + testCompanyID.String(),
			},
		},
		{
			name: "a user review evicts the farmer's average, not a company's list",
			req:  dto.CreateReviewRequest{TargetType: "user", TargetID: testOtherID, Rating: 4, TransactionID: txTestTransactID},
			wantKeys: []string{
				"reviews:byuser:" + testUserID.String(),
				"reviews:avg:user:" + testOtherID.String(),
			},
			// A user review has no company, so evicting bycompany would key
			// on the zero uuid.
			unwantedKeys: []string{"reviews:bycompany:" + uuid.Nil.String()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := newFakeCache()
			uc := usecases.NewCachedReviewUseCase(newTestReviewUC(newFakeReviewRepo()), cache)

			if _, err := uc.CreateReview(principalCtx(), tt.req); err != nil {
				t.Fatalf("CreateReview() error: %v", err)
			}

			for _, want := range tt.wantKeys {
				if !slices.Contains(cache.deletedKeys, want) {
					t.Errorf("evicted keys = %v, want it to contain %q", cache.deletedKeys, want)
				}
			}
			for _, unwanted := range tt.unwantedKeys {
				if slices.Contains(cache.deletedKeys, unwanted) {
					t.Errorf("evicted keys = %v, want it NOT to contain %q", cache.deletedKeys, unwanted)
				}
			}
		})
	}
}

// The aggregate is cached under its own key rather than recomputed on every
// profile view or smuggled through one of the list keys.
func TestReviewCacheReadsTheAggregateThroughItself(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	uc := usecases.NewCachedReviewUseCase(newTestReviewUC(newFakeReviewRepo()), cache)

	got, err := uc.GetAverageRating(context.Background(), domain.ReviewTargetCompany, testCompanyID)
	if err != nil {
		t.Fatalf("GetAverageRating() error: %v", err)
	}
	if got.Average != 4.5 || got.Count != 2 {
		t.Errorf("average/count = %v/%d, want 4.5/2", got.Average, got.Count)
	}

	want := "reviews:avg:company:" + testCompanyID.String()
	if !slices.Contains(cache.remembered, want) {
		t.Errorf("remembered keys = %v, want it to contain %q", cache.remembered, want)
	}
}
