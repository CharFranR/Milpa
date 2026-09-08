package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestOfferingUseCaseCreateOffering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		req        dto.CreateOfferingRequest
		companyErr error
		ownerID    uuid.UUID
		saveErr    error
		wantErr    error
		wantPrice  float64
	}{
		{
			name:      "happy path",
			ctx:       principalCtx(),
			req:       dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn", Description: "Fresh corn", Price: 10.5, ImageURL: "http://img.milpa.com/corn.png"},
			wantPrice: 10.5,
		},
		{
			name: "zero price ignored",
			ctx:  principalCtx(),
			req:  dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingService, Name: "Delivery"},
		},
		{name: "unauthenticated", ctx: context.Background(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn"}, wantErr: auth.ErrUnauthenticated},
		{name: "company error", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn"}, companyErr: errFake, wantErr: errFake},
		{name: "company not found", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn"}, companyErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "forbidden", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn"}, ownerID: testOtherID, wantErr: domain.ErrForbidden},
		{name: "empty name", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct}, wantErr: domain.ErrNameRequired},
		{name: "invalid type", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingType(99), Name: "Organic Corn"}, wantErr: domain.ErrInvalidOfferingType},
		{name: "save error", ctx: principalCtx(), req: dto.CreateOfferingRequest{CompanyID: testCompanyID, Type: domain.OfferingProduct, Name: "Organic Corn"}, saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyRepo := newFakeCompanyRepo()
			if tt.companyErr != nil || tt.ownerID != uuid.Nil {
				companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
					if tt.companyErr != nil {
						return nil, tt.companyErr
					}
					company := mustCompany()
					company.Owner.ID = tt.ownerID
					return company, nil
				}
			}
			offeringRepo := newFakeOfferingRepo()
			if tt.saveErr != nil {
				offeringRepo.save = func(ctx context.Context, offering *domain.Offering) error {
					return tt.saveErr
				}
			}
			uc := usecases.NewOfferingUseCase(offeringRepo, companyRepo, newFakeTimer())

			got, err := uc.CreateOffering(tt.ctx, tt.req)

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
			if got.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if got.CompanyID != tt.req.CompanyID {
				t.Errorf("company id = %v, want %v", got.CompanyID, tt.req.CompanyID)
			}
			if got.Type != tt.req.Type {
				t.Errorf("type = %v, want %v", got.Type, tt.req.Type)
			}
			if got.Name != tt.req.Name {
				t.Errorf("name = %q, want %q", got.Name, tt.req.Name)
			}
			if got.Description != tt.req.Description {
				t.Errorf("description = %q, want %q", got.Description, tt.req.Description)
			}
			if got.Price != tt.wantPrice {
				t.Errorf("price = %v, want %v", got.Price, tt.wantPrice)
			}
			if got.ImageURL != tt.req.ImageURL {
				t.Errorf("image url = %q, want %q", got.ImageURL, tt.req.ImageURL)
			}
			if !got.CreatedAt.Equal(fixedTime) || !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, fixedTime)
			}
			if len(offeringRepo.saved) != 1 {
				t.Fatalf("saved offerings = %d, want 1", len(offeringRepo.saved))
			}
			if saved := offeringRepo.saved[0]; saved.Price != tt.wantPrice {
				t.Errorf("saved price = %v, want %v", saved.Price, tt.wantPrice)
			}
		})
	}
}

func TestOfferingUseCaseGetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "happy path"},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
		{name: "not found", repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			if tt.repoErr != nil {
				offeringRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeCompanyRepo(), newFakeTimer())

			got, err := uc.GetByID(context.Background(), testOfferingID)

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
			if got.ID != testOfferingID {
				t.Errorf("id = %v, want %v", got.ID, testOfferingID)
			}
			if got.CompanyID != testCompanyID {
				t.Errorf("company id = %v, want %v", got.CompanyID, testCompanyID)
			}
			if got.Type != domain.OfferingProduct {
				t.Errorf("type = %v, want %v", got.Type, domain.OfferingProduct)
			}
			if got.Name != "Organic Corn" {
				t.Errorf("name = %q, want %q", got.Name, "Organic Corn")
			}
			if got.Description != "Fresh organic corn" {
				t.Errorf("description = %q, want %q", got.Description, "Fresh organic corn")
			}
			if got.Price != 10.0 {
				t.Errorf("price = %v, want 10", got.Price)
			}
			if got.ImageURL != "http://images.milpa.com/corn.png" {
				t.Errorf("image url = %q, want %q", got.ImageURL, "http://images.milpa.com/corn.png")
			}
			if !got.CreatedAt.Equal(fixedTime) || !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestOfferingUseCaseGetByCompany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		offerings []domain.Offering
		repoErr   error
		wantLen   int
		wantErr   error
	}{
		{
			name: "happy path",
			offerings: func() []domain.Offering {
				first := *mustOffering()
				second := *mustOffering()
				second.ID = testOtherID
				second.Name = "Honey"
				return []domain.Offering{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", offerings: []domain.Offering{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			if tt.repoErr != nil {
				offeringRepo.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error) {
					return nil, tt.repoErr
				}
			} else {
				offeringRepo.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error) {
					return tt.offerings, nil
				}
			}
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeCompanyRepo(), newFakeTimer())

			got, err := uc.GetByCompany(context.Background(), testCompanyID)

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
			if len(got) != tt.wantLen {
				t.Fatalf("dtos = %d, want %d", len(got), tt.wantLen)
			}
			for i, dto := range got {
				if dto.ID != tt.offerings[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.offerings[i].ID)
				}
				if dto.Name != tt.offerings[i].Name {
					t.Errorf("dto %d name = %q, want %q", i, dto.Name, tt.offerings[i].Name)
				}
				if dto.CompanyID != tt.offerings[i].CompanyID {
					t.Errorf("dto %d company id = %v, want %v", i, dto.CompanyID, tt.offerings[i].CompanyID)
				}
			}
		})
	}
}

func TestOfferingUseCaseUpdateOffering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       dto.UpdateOfferingRequest
		repoErr   error
		wantErr   error
		wantType  domain.OfferingType
		wantName  string
		wantDesc  string
		wantPrice float64
		wantImage string
	}{
		{name: "no fields", req: dto.UpdateOfferingRequest{}, wantType: domain.OfferingProduct, wantName: "Organic Corn", wantDesc: "Fresh organic corn", wantPrice: 10.0, wantImage: "http://images.milpa.com/corn.png"},
		{
			name:      "all fields",
			req:       dto.UpdateOfferingRequest{Type: offeringTypePtr(domain.OfferingService), Name: strPtr("Delivery"), Description: strPtr("Fast delivery"), Price: floatPtr(25.0), ImageURL: strPtr("http://img.milpa.com/delivery.png")},
			wantType:  domain.OfferingService,
			wantName:  "Delivery",
			wantDesc:  "Fast delivery",
			wantPrice: 25.0,
			wantImage: "http://img.milpa.com/delivery.png",
		},
		{name: "type only", req: dto.UpdateOfferingRequest{Type: offeringTypePtr(domain.OfferingService)}, wantType: domain.OfferingService, wantName: "Organic Corn", wantDesc: "Fresh organic corn", wantPrice: 10.0, wantImage: "http://images.milpa.com/corn.png"},
		{name: "name only", req: dto.UpdateOfferingRequest{Name: strPtr("Delivery")}, wantType: domain.OfferingProduct, wantName: "Delivery", wantDesc: "Fresh organic corn", wantPrice: 10.0, wantImage: "http://images.milpa.com/corn.png"},
		{name: "description only", req: dto.UpdateOfferingRequest{Description: strPtr("Fast delivery")}, wantType: domain.OfferingProduct, wantName: "Organic Corn", wantDesc: "Fast delivery", wantPrice: 10.0, wantImage: "http://images.milpa.com/corn.png"},
		{name: "price only", req: dto.UpdateOfferingRequest{Price: floatPtr(25.0)}, wantType: domain.OfferingProduct, wantName: "Organic Corn", wantDesc: "Fresh organic corn", wantPrice: 25.0, wantImage: "http://images.milpa.com/corn.png"},
		{name: "image only", req: dto.UpdateOfferingRequest{ImageURL: strPtr("http://img.milpa.com/delivery.png")}, wantType: domain.OfferingProduct, wantName: "Organic Corn", wantDesc: "Fresh organic corn", wantPrice: 10.0, wantImage: "http://img.milpa.com/delivery.png"},
		{name: "invalid price", req: dto.UpdateOfferingRequest{Price: floatPtr(0)}, wantErr: domain.ErrInvalidPrice},
		{name: "repo error", req: dto.UpdateOfferingRequest{Name: strPtr("Delivery")}, repoErr: errFake, wantErr: errFake},
		{name: "not found", req: dto.UpdateOfferingRequest{Name: strPtr("Delivery")}, repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			if tt.repoErr != nil {
				offeringRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeCompanyRepo(), newFakeTimer())

			err := uc.UpdateOffering(context.Background(), testOfferingID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if tt.wantErr == domain.ErrInvalidPrice && len(offeringRepo.updated) != 0 {
					t.Errorf("updated offerings = %d, want 0", len(offeringRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(offeringRepo.updated) != 1 {
				t.Fatalf("updated offerings = %d, want 1", len(offeringRepo.updated))
			}
			updated := offeringRepo.updated[0]
			if updated.Type != tt.wantType {
				t.Errorf("type = %v, want %v", updated.Type, tt.wantType)
			}
			if updated.Name != tt.wantName {
				t.Errorf("name = %q, want %q", updated.Name, tt.wantName)
			}
			if updated.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", updated.Description, tt.wantDesc)
			}
			if updated.Price != tt.wantPrice {
				t.Errorf("price = %v, want %v", updated.Price, tt.wantPrice)
			}
			if updated.ImageURL != tt.wantImage {
				t.Errorf("image url = %q, want %q", updated.ImageURL, tt.wantImage)
			}
			if !updated.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", updated.UpdatedAt, fixedTime)
			}
		})
	}
}
