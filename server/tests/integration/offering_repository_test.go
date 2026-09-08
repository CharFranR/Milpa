package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
)

var testOfferingID uuid.UUID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
var testOfferingID2 uuid.UUID = uuid.MustParse("44444444-4444-4444-4444-444444444445")
var testOfferingCompanyID uuid.UUID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

func setupOfferingTestData(t *testing.T) {
	t.Helper()

	userRepo := repository.NewUserRepository(TestPool)
	companyRepo := repository.NewCompanyRepository(TestPool)

	_, err := userRepo.Save(context.Background(), &domain.User{
		ID:           testOwnerID,
		FirstName:    "Owner",
		LastName:     "User",
		Role:         domain.RoleMIPYME,
		Email:        "offering-owner@example.com",
		PhoneNumber:  "0000-0000",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	})
	if err != nil {
		t.Fatalf("insert owner user: %v", err)
	}

	err = companyRepo.Save(context.Background(), &domain.Company{
		ID:          testOfferingCompanyID,
		Name:        "Offering Company",
		Owner:       domain.User{ID: testOwnerID},
		Address:     domain.Address{},
		Description: "Company for offerings",
		PhoneNumber: "1234-5678",
		Email:       "offering-company@example.com",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	})
	if err != nil {
		t.Fatalf("insert company: %v", err)
	}
}

func TestOfferingSave(t *testing.T) {
	setupOfferingTestData(t)
	db := repository.NewOfferingRepository(TestPool)

	tests := []struct {
		Name        string
		ExpectedErr error
		Offering    *domain.Offering
	}{
		{
			Name: "Happy Path product",
			Offering: &domain.Offering{
				ID:          testOfferingID,
				CompanyID:   testOfferingCompanyID,
				Type:        domain.OfferingProduct,
				Name:        "Laptop",
				Description: "Gaming laptop",
				Price:       1200.50,
				ImageURL:    "http://img",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			Name: "Happy Path service",
			Offering: &domain.Offering{
				ID:          testOfferingID2,
				CompanyID:   testOfferingCompanyID,
				Type:        domain.OfferingService,
				Name:        "Consultoria",
				Description: "IT consulting",
				Price:       100.00,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Save(context.Background(), tt.Offering)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Save() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Save() unexpected error: %v", err)
			}
		})
	}
}

func TestOfferingFindByID(t *testing.T) {
	setupOfferingTestData(t)
	db := repository.NewOfferingRepository(TestPool)

	saved := &domain.Offering{
		ID:          testOfferingID,
		CompanyID:   testOfferingCompanyID,
		Type:        domain.OfferingProduct,
		Name:        "Laptop",
		Description: "Gaming laptop",
		Price:       1200.50,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testOfferingID,
			ExpectedErr: nil,
		},
		{
			Name:        "Not Found",
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444499"),
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			offering, err := db.FindByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByID() unexpected error: %v", err)
				return
			}

			if offering == nil {
				t.Fatal("FindByID() returned nil offering")
			}
			if offering.ID != tt.ID {
				t.Errorf("FindByID() ID = %v, want %v", offering.ID, tt.ID)
			}
			if offering.Name != saved.Name {
				t.Errorf("FindByID() Name = %v, want %v", offering.Name, saved.Name)
			}
			if offering.Price != saved.Price {
				t.Errorf("FindByID() Price = %v, want %v", offering.Price, saved.Price)
			}
		})
	}
}

func TestOfferingFindByCompany(t *testing.T) {
	setupOfferingTestData(t)
	db := repository.NewOfferingRepository(TestPool)

	offering1 := &domain.Offering{
		ID:        testOfferingID,
		CompanyID: testOfferingCompanyID,
		Type:      domain.OfferingProduct,
		Name:      "Laptop",
		Price:     1200.50,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	offering2 := &domain.Offering{
		ID:        testOfferingID2,
		CompanyID: testOfferingCompanyID,
		Type:      domain.OfferingService,
		Name:      "Consultoria",
		Price:     100.00,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	if err := db.Save(context.Background(), offering1); err != nil {
		t.Fatalf("Save offering1: %v", err)
	}
	if err := db.Save(context.Background(), offering2); err != nil {
		t.Fatalf("Save offering2: %v", err)
	}

	tests := []struct {
		Name          string
		CompanyID     uuid.UUID
		ExpectedLen   int
		ExpectedNames []string
	}{
		{
			Name:          "Company with offerings",
			CompanyID:     testOfferingCompanyID,
			ExpectedLen:   2,
			ExpectedNames: []string{"Laptop", "Consultoria"},
		},
		{
			Name:        "Company with no offerings",
			CompanyID:   uuid.MustParse("33333333-3333-3333-3333-333333333399"),
			ExpectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			offerings, err := db.FindByCompany(context.Background(), tt.CompanyID)

			if err != nil {
				t.Errorf("FindByCompany() unexpected error: %v", err)
				return
			}

			if len(offerings) < tt.ExpectedLen {
				t.Fatalf("FindByCompany() got %d offerings, want at least %d", len(offerings), tt.ExpectedLen)
			}

			if tt.ExpectedNames != nil {
				found := make(map[string]bool)
				for _, o := range offerings {
					found[o.Name] = true
				}
				for _, name := range tt.ExpectedNames {
					if !found[name] {
						t.Errorf("FindByCompany() did not find offering '%s'", name)
					}
				}
			}
		})
	}
}

func TestOfferingUpdate(t *testing.T) {
	setupOfferingTestData(t)
	db := repository.NewOfferingRepository(TestPool)

	saved := &domain.Offering{
		ID:          testOfferingID,
		CompanyID:   testOfferingCompanyID,
		Type:        domain.OfferingProduct,
		Name:        "Laptop",
		Description: "Original desc",
		Price:       1200.50,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ExpectedErr error
		update_func func() *domain.Offering
	}{
		{
			Name: "Happy path",
			update_func: func() *domain.Offering {
				return &domain.Offering{
					ID:          testOfferingID,
					CompanyID:   testOfferingCompanyID,
					Type:        domain.OfferingProduct,
					Name:        "Laptop Pro",
					Description: "Updated desc",
					Price:       1500.00,
					ImageURL:    "http://new-img",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime2,
				}
			},
		},
		{
			Name: "Update name only",
			update_func: func() *domain.Offering {
				return &domain.Offering{
					ID:          testOfferingID,
					CompanyID:   testOfferingCompanyID,
					Type:        domain.OfferingProduct,
					Name:        "Laptop Ultra",
					Description: "Updated desc",
					Price:       1500.00,
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime2,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Update(context.Background(), tt.update_func())

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
			}
		})
	}
}

func TestOfferingDelete(t *testing.T) {
	setupOfferingTestData(t)
	db := repository.NewOfferingRepository(TestPool)

	saved := &domain.Offering{
		ID:        testOfferingID,
		CompanyID: testOfferingCompanyID,
		Type:      domain.OfferingProduct,
		Name:      "To Delete",
		Price:     50.00,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy path",
			ID:          testOfferingID,
			ExpectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Delete(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Delete() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Delete() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error: %v", err)
			}

			_, findErr := db.FindByID(context.Background(), tt.ID)
			if findErr == nil {
				t.Error("FindByID() after Delete() should return error, got nil")
			}
		})
	}
}
