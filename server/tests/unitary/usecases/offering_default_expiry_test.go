package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

// TestOfferingCreateExpiryFallback characterizes the candidate's default-expiry
// behavior: when the request omits expires_at, the offering inherits
// now + default_expiry_days from its category. An explicit expires_at always
// wins, and a category without a default (nil or zero) leaves the offering
// without an expiry instead of inventing one.
func TestOfferingCreateExpiryFallback(t *testing.T) {
	t.Parallel()

	explicit := fixedTime.Add(48 * time.Hour)
	tenDays := 10
	zeroDays := 0
	expectedDefault := fixedTime.AddDate(0, 0, tenDays)

	tests := []struct {
		name        string
		explicit    *time.Time
		category    func() *domain.Category
		categoryErr error
		wantExpiry  *time.Time
		wantErr     error
	}{
		{
			name:       "category default fills an omitted expiry",
			category:   categoryWithDefaultExpiry(&tenDays),
			wantExpiry: &expectedDefault,
		},
		{
			name:       "explicit expiry beats the category default",
			explicit:   &explicit,
			category:   categoryWithDefaultExpiry(&tenDays),
			wantExpiry: &explicit,
		},
		{
			name:       "nil category default means no expiry",
			category:   categoryWithDefaultExpiry(nil),
			wantExpiry: nil,
		},
		{
			name:       "zero category default means no expiry",
			category:   categoryWithDefaultExpiry(&zeroDays),
			wantExpiry: nil,
		},
		{
			name:        "unknown category is not found",
			categoryErr: domain.ErrNotFound,
			wantErr:     domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			categoryRepo := newFakeCategoryRepo()
			categoryRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
				if tt.categoryErr != nil {
					return nil, tt.categoryErr
				}
				category := tt.category()
				category.ID = id
				return category, nil
			}

			offeringRepo := newFakeOfferingRepo()
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), categoryRepo, newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			req := completeCatalogueRequest()
			req.ExpiresAt = tt.explicit

			_, err := uc.CreateOffering(farmerCtx(), req)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CreateOffering() error = %v, want %v", err, tt.wantErr)
				}
				if len(offeringRepo.saved) != 0 {
					t.Errorf("a refused create saved %d offerings, want 0", len(offeringRepo.saved))
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateOffering() error: %v", err)
			}
			if len(offeringRepo.saved) != 1 {
				t.Fatalf("saved = %d offerings, want 1", len(offeringRepo.saved))
			}

			got := offeringRepo.saved[0].ExpiresAt
			switch {
			case tt.wantExpiry == nil && got != nil:
				t.Errorf("expires at = %v, want no expiry", got)
			case tt.wantExpiry != nil && got == nil:
				t.Errorf("expires at = nil, want %v", tt.wantExpiry)
			case tt.wantExpiry != nil && !got.Equal(*tt.wantExpiry):
				t.Errorf("expires at = %v, want %v", got, tt.wantExpiry)
			}
		})
	}
}

func categoryWithDefaultExpiry(days *int) func() *domain.Category {
	return func() *domain.Category {
		category := mustCategory()
		category.DefaultExpiryDays = days
		return category
	}
}
