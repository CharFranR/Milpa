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

func TestCompanyUseCaseCreateCompany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		req        dto.RegisterCompanyRequest
		categoryID uuid.UUID
		catErr     error
		saveErr    error
		wantErr    error
	}{
		{
			name: "happy path",
			ctx:  principalCtx(),
			req:  dto.RegisterCompanyRequest{Name: "Milpa S.A.", Address: "Managua", Description: "Organic farm", PhoneNumber: "555-0000", Email: "info@milpa.com", Website: "milpa.com"},
		},
		{
			name:       "with category",
			ctx:        principalCtx(),
			req:        dto.RegisterCompanyRequest{Name: "Milpa S.A.", CategoryID: testCategoryID},
			categoryID: testCategoryID,
		},
		{name: "unauthenticated", ctx: context.Background(), req: dto.RegisterCompanyRequest{Name: "Milpa S.A."}, wantErr: auth.ErrUnauthenticated},
		{name: "owner required", ctx: auth.WithPrincipal(context.Background(), auth.Principal{}), req: dto.RegisterCompanyRequest{Name: "Milpa S.A."}, wantErr: domain.ErrOwnerRequired},
		{name: "empty name", ctx: principalCtx(), req: dto.RegisterCompanyRequest{}, wantErr: domain.ErrNameRequired},
		{name: "category error", ctx: principalCtx(), req: dto.RegisterCompanyRequest{Name: "Milpa S.A.", CategoryID: testCategoryID}, catErr: errFake, wantErr: errFake},
		{name: "category not found", ctx: principalCtx(), req: dto.RegisterCompanyRequest{Name: "Milpa S.A.", CategoryID: testCategoryID}, catErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "save error", ctx: principalCtx(), req: dto.RegisterCompanyRequest{Name: "Milpa S.A."}, saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyRepo := newFakeCompanyRepo()
			if tt.saveErr != nil {
				companyRepo.save = func(ctx context.Context, company *domain.Company) error {
					return tt.saveErr
				}
			}
			categoryRepo := newFakeCategoryRepo()
			if tt.catErr != nil {
				categoryRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
					return nil, tt.catErr
				}
			}
			uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), categoryRepo, newFakeTimer())

			got, err := uc.CreateCompany(tt.ctx, tt.req)

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
			if got.Name != tt.req.Name {
				t.Errorf("name = %q, want %q", got.Name, tt.req.Name)
			}
			if got.OwnerID != testUserID {
				t.Errorf("owner id = %v, want %v", got.OwnerID, testUserID)
			}
			if got.CategoryID != tt.categoryID {
				t.Errorf("category id = %v, want %v", got.CategoryID, tt.categoryID)
			}
			if got.Address != tt.req.Address+", , " {
				t.Errorf("address = %q, want %q", got.Address, tt.req.Address+", , ")
			}
			if got.Description != tt.req.Description {
				t.Errorf("description = %q, want %q", got.Description, tt.req.Description)
			}
			if got.PhoneNumber != tt.req.PhoneNumber {
				t.Errorf("phone number = %q, want %q", got.PhoneNumber, tt.req.PhoneNumber)
			}
			if got.Email != tt.req.Email {
				t.Errorf("email = %q, want %q", got.Email, tt.req.Email)
			}
			if got.Website != tt.req.Website {
				t.Errorf("website = %q, want %q", got.Website, tt.req.Website)
			}
			if got.Verified {
				t.Error("verified = true, want false")
			}
			if !got.CreatedAt.Equal(fixedTime) || !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, fixedTime)
			}
			if len(companyRepo.saved) != 1 {
				t.Fatalf("saved companies = %d, want 1", len(companyRepo.saved))
			}
			saved := companyRepo.saved[0]
			if tt.categoryID != uuid.Nil {
				if len(saved.Category) != 1 || saved.Category[0].ID != tt.categoryID {
					t.Errorf("saved category = %+v, want one category with id %v", saved.Category, tt.categoryID)
				}
			} else if len(saved.Category) != 0 {
				t.Errorf("saved category = %+v, want none", saved.Category)
			}
		})
	}
}

func TestCompanyUseCaseGetByID(t *testing.T) {
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

			companyRepo := newFakeCompanyRepo()
			if tt.repoErr != nil {
				companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer())

			got, err := uc.GetByID(context.Background(), testCompanyID)

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
			if got.ID != testCompanyID {
				t.Errorf("id = %v, want %v", got.ID, testCompanyID)
			}
			if got.Name != "Milpa S.A." {
				t.Errorf("name = %q, want %q", got.Name, "Milpa S.A.")
			}
			if got.OwnerID != testUserID {
				t.Errorf("owner id = %v, want %v", got.OwnerID, testUserID)
			}
			if got.CategoryID != uuid.Nil {
				t.Errorf("category id = %v, want nil", got.CategoryID)
			}
			if got.Verified {
				t.Error("verified = true, want false")
			}
			if !got.CreatedAt.Equal(fixedTime) || !got.UpdatedAt.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestCompanyUseCaseGetByOwner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		companies []domain.Company
		repoErr   error
		wantLen   int
		wantErr   error
	}{
		{
			name: "happy path",
			companies: func() []domain.Company {
				first := *mustCompany()
				second := *mustCompany()
				second.ID = testOtherID
				second.Name = "Finca Verde"
				return []domain.Company{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", companies: []domain.Company{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyRepo := newFakeCompanyRepo()
			if tt.repoErr != nil {
				companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
					return nil, tt.repoErr
				}
			} else {
				companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
					return tt.companies, nil
				}
			}
			uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer())

			got, err := uc.GetByOwner(context.Background(), testUserID)

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
				if dto.ID != tt.companies[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.companies[i].ID)
				}
				if dto.Name != tt.companies[i].Name {
					t.Errorf("dto %d name = %q, want %q", i, dto.Name, tt.companies[i].Name)
				}
				if dto.OwnerID != tt.companies[i].Owner.ID {
					t.Errorf("dto %d owner id = %v, want %v", i, dto.OwnerID, tt.companies[i].Owner.ID)
				}
			}
		})
	}
}

func TestCompanyUseCaseUpdateCompany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		req             dto.UpdateCompanyRequest
		repoErr         error
		wantErr         error
		wantName        string
		wantAddressLine string
		wantDescription string
		wantPhone       string
		wantEmail       string
		wantWebsite     string
	}{
		{
			name:            "no fields",
			req:             dto.UpdateCompanyRequest{},
			wantName:        "Milpa S.A.",
			wantDescription: "Organic farm",
			wantEmail:       "info@milpa.com",
		},
		{
			name: "all fields",
			req: dto.UpdateCompanyRequest{
				Name:        strPtr("Milpa Green"),
				Address:     strPtr("Matagalpa"),
				Description: strPtr("Certified organic"),
				PhoneNumber: strPtr("888-1111"),
				Email:       strPtr("hello@milpa.com"),
				Website:     strPtr("milpa.com.ni"),
			},
			wantName:        "Milpa Green",
			wantAddressLine: "Matagalpa",
			wantDescription: "Certified organic",
			wantPhone:       "888-1111",
			wantEmail:       "hello@milpa.com",
			wantWebsite:     "milpa.com.ni",
		},
		{name: "name only", req: dto.UpdateCompanyRequest{Name: strPtr("Milpa Green")}, wantName: "Milpa Green", wantDescription: "Organic farm", wantEmail: "info@milpa.com"},
		{name: "address only", req: dto.UpdateCompanyRequest{Address: strPtr("Matagalpa")}, wantName: "Milpa S.A.", wantAddressLine: "Matagalpa", wantDescription: "Organic farm", wantEmail: "info@milpa.com"},
		{name: "description only", req: dto.UpdateCompanyRequest{Description: strPtr("Certified organic")}, wantName: "Milpa S.A.", wantDescription: "Certified organic", wantEmail: "info@milpa.com"},
		{name: "phone only", req: dto.UpdateCompanyRequest{PhoneNumber: strPtr("888-1111")}, wantName: "Milpa S.A.", wantDescription: "Organic farm", wantEmail: "info@milpa.com", wantPhone: "888-1111"},
		{name: "email only", req: dto.UpdateCompanyRequest{Email: strPtr("hello@milpa.com")}, wantName: "Milpa S.A.", wantDescription: "Organic farm", wantEmail: "hello@milpa.com"},
		{name: "website only", req: dto.UpdateCompanyRequest{Website: strPtr("milpa.com.ni")}, wantName: "Milpa S.A.", wantDescription: "Organic farm", wantEmail: "info@milpa.com", wantWebsite: "milpa.com.ni"},
		{name: "repo error", req: dto.UpdateCompanyRequest{Name: strPtr("Milpa Green")}, repoErr: errFake, wantErr: errFake},
		{name: "not found", req: dto.UpdateCompanyRequest{Name: strPtr("Milpa Green")}, repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyRepo := newFakeCompanyRepo()
			if tt.repoErr != nil {
				companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
					return nil, tt.repoErr
				}
			} else {
				companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
					company := mustCompany()
					company.Description = "Organic farm"
					company.Email = "info@milpa.com"
					return company, nil
				}
			}
			uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer())

			err := uc.UpdateCompany(context.Background(), testCompanyID, tt.req)

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
			if len(companyRepo.updated) != 1 {
				t.Fatalf("updated companies = %d, want 1", len(companyRepo.updated))
			}
			updated := companyRepo.updated[0]
			if updated.Name != tt.wantName {
				t.Errorf("name = %q, want %q", updated.Name, tt.wantName)
			}
			if updated.Address.AddressLine != tt.wantAddressLine {
				t.Errorf("address line = %q, want %q", updated.Address.AddressLine, tt.wantAddressLine)
			}
			if updated.Description != tt.wantDescription {
				t.Errorf("description = %q, want %q", updated.Description, tt.wantDescription)
			}
			if updated.PhoneNumber != tt.wantPhone {
				t.Errorf("phone number = %q, want %q", updated.PhoneNumber, tt.wantPhone)
			}
			if updated.Email != tt.wantEmail {
				t.Errorf("email = %q, want %q", updated.Email, tt.wantEmail)
			}
			if updated.Website != tt.wantWebsite {
				t.Errorf("website = %q, want %q", updated.Website, tt.wantWebsite)
			}
			if !updated.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", updated.UpdatedAt, fixedTime)
			}
		})
	}
}
