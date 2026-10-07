package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestOfferingCreateRefusesAnIncompleteProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		blank   func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest
		wantErr error
	}{
		{
			name: "no variety",
			blank: func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest {
				req.Variety = ""
				return req
			},
			wantErr: domain.ErrVarietyRequired,
		},
		{
			name: "no unit of measure",
			blank: func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest {
				req.UnitOfMeasureID = nil
				return req
			},
			wantErr: domain.ErrUnitOfMeasureRequired,
		},
		{
			name: "no available quantity",
			blank: func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest {
				req.QuantityAvailable = 0
				return req
			},
			wantErr: domain.ErrInvalidQuantity,
		},
		{
			name: "negative available quantity",
			blank: func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest {
				req.QuantityAvailable = -5
				return req
			},
			wantErr: domain.ErrInvalidQuantity,
		},
		{
			name: "no category",
			blank: func(req dto.CreateOfferingRequest) dto.CreateOfferingRequest {
				req.CategoryID = nil
				return req
			},
			wantErr: domain.ErrCategoryRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			_, err := uc.CreateOffering(farmerCtx(), tt.blank(completeCatalogueRequest()))

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateOffering() error = %v, want %v", err, tt.wantErr)
			}
			if len(offeringRepo.saved) != 0 {
				t.Errorf("an incomplete product was persisted %d times, want 0", len(offeringRepo.saved))
			}
		})
	}
}

func TestOfferingUpdateRefusesAnIncompleteProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		blank   func(req dto.UpdateOfferingRequest) dto.UpdateOfferingRequest
		wantErr error
	}{
		{
			name: "no variety",
			blank: func(req dto.UpdateOfferingRequest) dto.UpdateOfferingRequest {
				req.Variety = strPtr("")
				return req
			},
			wantErr: domain.ErrVarietyRequired,
		},
		{
			name: "no available quantity",
			blank: func(req dto.UpdateOfferingRequest) dto.UpdateOfferingRequest {
				req.QuantityAvailable = floatPtr(0)
				return req
			},
			wantErr: domain.ErrInvalidQuantity,
		},
		{
			name: "half a coordinate pair",
			blank: func(req dto.UpdateOfferingRequest) dto.UpdateOfferingRequest {
				latitude := 12.4379
				req.Latitude = &latitude
				return req
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "coordinates off the globe",
			blank: func(req dto.UpdateOfferingRequest) dto.UpdateOfferingRequest {
				latitude := 120.0
				longitude := -86.0
				req.Latitude = &latitude
				req.Longitude = &longitude
				return req
			},
			wantErr: domain.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			err := uc.UpdateOffering(farmerCtx(), testOfferingID, tt.blank(dto.UpdateOfferingRequest{}))

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateOffering() error = %v, want %v", err, tt.wantErr)
			}
			if len(offeringRepo.updated) != 0 {
				t.Errorf("an incomplete product was persisted %d times, want 0", len(offeringRepo.updated))
			}
		})
	}
}

func TestOfferingCreateStoresEveryCatalogueField(t *testing.T) {
	t.Parallel()

	offeringRepo := newFakeOfferingRepo()
	uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

	expiresAt := fixedTime.Add(72 * time.Hour)
	companyID := testCompanyID
	latitude := 12.4379
	longitude := -86.8781

	req := completeCatalogueRequest()
	req.ExpiresAt = &expiresAt
	req.CompanyID = &companyID
	req.Latitude = &latitude
	req.Longitude = &longitude

	got, err := uc.CreateOffering(farmerCtx(), req)
	if err != nil {
		t.Fatalf("CreateOffering() error: %v", err)
	}

	if got.Variety != req.Variety {
		t.Errorf("variety = %q, want %q", got.Variety, req.Variety)
	}
	if got.UnitOfMeasureID == nil || *got.UnitOfMeasureID != testUnitOfMeasureID {
		t.Errorf("unit of measure = %v, want %v", got.UnitOfMeasureID, testUnitOfMeasureID)
	}
	if got.QuantityAvailable != req.QuantityAvailable {
		t.Errorf("quantity = %v, want %v", got.QuantityAvailable, req.QuantityAvailable)
	}
	if got.CategoryID == nil || *got.CategoryID != testCategoryID {
		t.Errorf("category = %v, want %v", got.CategoryID, testCategoryID)
	}
	if got.CompanyID == nil || *got.CompanyID != companyID {
		t.Errorf("company = %v, want %v", got.CompanyID, companyID)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expires at = %v, want %v", got.ExpiresAt, expiresAt)
	}
	if got.Latitude == nil || *got.Latitude != latitude {
		t.Errorf("latitude = %v, want %v", got.Latitude, latitude)
	}
	if got.Longitude == nil || *got.Longitude != longitude {
		t.Errorf("longitude = %v, want %v", got.Longitude, longitude)
	}
	if !got.IsActive {
		t.Error("a newly published product is inactive")
	}
}

func TestOfferingDeactivateIsIdempotentAndOwnerOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		wantErr     error
		wantUpdates int
	}{
		{name: "unauthenticated", ctx: context.Background(), wantErr: errFake},
		{name: "foreign user", ctx: farmerCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "owner", ctx: farmerCtx(), wantUpdates: 1},
		{name: "admin cannot pass the farmer guard", ctx: reportAdminCtx(), wantErr: domain.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			var err error
			if tt.ctx == nil {
				_, err = uc.DeactivateOffering(context.Background(), testOfferingID)
			} else {
				_, err = uc.DeactivateOffering(tt.ctx, testOfferingID)
			}

			if tt.wantErr == errFake {
				if err == nil {
					t.Fatal("DeactivateOffering() error = nil, want an error for an anonymous caller")
				}
				return
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("DeactivateOffering() error = %v, want %v", err, tt.wantErr)
				}
				if len(offeringRepo.updated) != 0 {
					t.Errorf("a refused caller wrote %d offerings, want 0", len(offeringRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("DeactivateOffering() error: %v", err)
			}
			if len(offeringRepo.updated) != tt.wantUpdates {
				t.Fatalf("updated = %d, want %d", len(offeringRepo.updated), tt.wantUpdates)
			}
			for _, offering := range offeringRepo.updated {
				if offering.IsActive {
					t.Error("the offering is still active after deactivation")
				}
			}
		})
	}
}

func TestOfferingRenewIsOwnerOnlyAndReactivatesTheProduct(t *testing.T) {
	t.Parallel()

	newExpiry := fixedTime.Add(15 * 24 * time.Hour)

	tests := []struct {
		name        string
		ctx         context.Context
		wantErr     error
		wantUpdates int
	}{
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "foreign user", ctx: farmerCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "owner", ctx: farmerCtx(), wantUpdates: 1},
		{name: "admin cannot pass the farmer guard", ctx: reportAdminCtx(), wantErr: domain.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			offeringRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
				offering := mustOffering()
				offering.ID = id
				offering.Deactivate()
				expiry := fixedTime.Add(-time.Hour)
				offering.ExpiresAt = &expiry
				return offering, nil
			}
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			result, err := uc.RenewOffering(tt.ctx, testOfferingID, dto.RenewOfferingRequest{ExpiresAt: newExpiry})

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RenewOffering() error = %v, want %v", err, tt.wantErr)
				}
				if len(offeringRepo.updated) != 0 {
					t.Errorf("a refused caller wrote %d offerings, want 0", len(offeringRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("RenewOffering() error: %v", err)
			}
			if len(offeringRepo.updated) != tt.wantUpdates {
				t.Fatalf("updated = %d, want %d", len(offeringRepo.updated), tt.wantUpdates)
			}
			if result.ExpiresAt == nil || !result.ExpiresAt.Equal(newExpiry) {
				t.Errorf("expires at = %v, want %v", result.ExpiresAt, newExpiry)
			}
			if !result.IsActive {
				t.Error("the renewed product is not active")
			}
			if result.UpdatedAt != fixedTime {
				t.Errorf("updated at = %v, want %v", result.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestIndexRequestPrefersTheProductLocation(t *testing.T) {
	t.Parallel()

	farmerLatitude := 12.4379
	farmerLongitude := -86.8781
	productLatitude := 11.9747
	productLongitude := -86.0941

	tests := []struct {
		name          string
		productLat    *float64
		productLon    *float64
		wantLatitude  float64
		wantLongitude float64
	}{
		{
			name:          "product location wins",
			productLat:    &productLatitude,
			productLon:    &productLongitude,
			wantLatitude:  productLatitude,
			wantLongitude: productLongitude,
		},
		{
			name:          "farmer location is the fallback",
			wantLatitude:  farmerLatitude,
			wantLongitude: farmerLongitude,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			indexed := &capturingFuzzyRetrival{}
			uc := usecases.NewOfferingUseCase(
				newFakeOfferingRepo(),
				farmerAt(farmerLatitude, farmerLongitude),
				newFakeCategoryRepo(),
				newFakeTimer(),
				indexed,
				&fakeInvalidator{},
			)

			req := completeCatalogueRequest()
			req.Latitude = tt.productLat
			req.Longitude = tt.productLon

			if _, err := uc.CreateOffering(farmerCtx(), req); err != nil {
				t.Fatalf("CreateOffering() error: %v", err)
			}

			if indexed.document == nil {
				t.Fatal("no search document was indexed")
			}
			if indexed.document.Latitude != tt.wantLatitude {
				t.Errorf("indexed latitude = %v, want %v", indexed.document.Latitude, tt.wantLatitude)
			}
			if indexed.document.Longitude != tt.wantLongitude {
				t.Errorf("indexed longitude = %v, want %v", indexed.document.Longitude, tt.wantLongitude)
			}
		})
	}
}
