package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

// TestOfferingGetByUserIDIncludeHiddenPolicy pins the visibility gate: asking
// for hidden offerings is only legal for the owner behind a token, because the
// public route would otherwise leak deactivated products.
func TestOfferingGetByUserIDIncludeHiddenPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ctx           context.Context
		wantErr       error
		wantForwarded bool
	}{
		{name: "no principal is unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "a different principal is forbidden", ctx: principalCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "the owner passes with includeHidden", ctx: principalCtx(), wantForwarded: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			forwarded := false
			offeringRepo.findByUserID = func(ctx context.Context, userID uuid.UUID) ([]domain.Offering, error) {
				forwarded = true
				return []domain.Offering{}, nil
			}

			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			_, err := uc.GetByUserID(tt.ctx, testUserID, true)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetByUserID() error = %v, want %v", err, tt.wantErr)
				}
				if forwarded {
					t.Error("a refused caller reached the repository")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByUserID() error: %v", err)
			}
			if !forwarded {
				t.Fatal("the repository was not asked for the owner's offerings")
			}
			if !offeringRepo.lastFindIncludeHidden {
				t.Error("includeHidden was not forwarded to the repository")
			}
		})
	}
}

// TestCachedOfferingGetByUserIDIncludeHiddenBypassesTheCache guards the other
// half: the owner's full list (hidden included) has no cache entry, so it must
// not read or write the shared offerings:byuser key that serves the public list.
func TestCachedOfferingGetByUserIDIncludeHiddenBypassesTheCache(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	uc := cachedOfferingUC(cache)

	if _, err := uc.GetByUserID(principalCtx(), testUserID, true); err != nil {
		t.Fatalf("GetByUserID(includeHidden=true) error: %v", err)
	}

	if cache.calledGet || cache.calledSet || len(cache.remembered) != 0 {
		t.Errorf("a hidden listing touched the cache: get=%v set=%v remembered=%v", cache.calledGet, cache.calledSet, cache.remembered)
	}
}
