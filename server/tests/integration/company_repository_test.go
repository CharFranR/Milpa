package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testCompanyID uuid.UUID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
var testCompanyID2 uuid.UUID = uuid.MustParse("33333333-3333-3333-3333-333333333334")
var testCategoryID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222201")
var testOwnerID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var fixedTime2 time.Time = time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)

func setupCompanyTestData(t *testing.T) {
	t.Helper()

	userRepo := repository.NewUserRepository(TestPool)
	catRepo := repository.NewCategoryRepository(TestPool)

	_, err := userRepo.Save(context.Background(), &domain.User{
		ID:           testOwnerID,
		FirstName:    "Owner",
		LastName:     "User",
		Role:         domain.RoleMIPYME,
		Email:        "owner@example.com",
		PhoneNumber:  "0000-0000",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	})
	if err != nil {
		t.Fatalf("insert owner user: %v", err)
	}

	err = catRepo.Save(context.Background(), &domain.Category{
		ID:          testCategoryID,
		Name:        "Tech",
		Description: "Technology",
	})
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
}

func TestCompanySave(t *testing.T) {
	setupCompanyTestData(t)
	db := repository.NewCompanyRepository(TestPool)

	cat := domain.Category{ID: testCategoryID, Name: "Tech", Description: "Technology"}

	tests := []struct {
		Name        string
		ExpectedErr error
		Company     *domain.Company
	}{
		{
			Name: "Happy Path with address",
			Company: &domain.Company{
				ID:          testCompanyID,
				Name:        "Mi Empresa",
				Owner:       domain.User{ID: testOwnerID},
				Category:    []domain.Category{cat},
				Address:     domain.Address{Department: "Leon", Municipality: "Leon", AddressLine: "Calle Central"},
				Description: "Una empresa de tecnologia",
				PhoneNumber: "1234-5678",
				Email:       "empresa@example.com",
				Website:     "http://web",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			Name: "Happy Path without address",
			Company: &domain.Company{
				ID:          testCompanyID2,
				Name:        "Otra Empresa",
				Owner:       domain.User{ID: testOwnerID},
				Category:    []domain.Category{cat},
				Address:     domain.Address{},
				Description: "Otra empresa",
				PhoneNumber: "9999-9999",
				Email:       "otra@example.com",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Save(context.Background(), tt.Company)

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

func TestCompanyFindByID(t *testing.T) {
	setupCompanyTestData(t)
	db := repository.NewCompanyRepository(TestPool)

	savedCompany := &domain.Company{
		ID:          testCompanyID,
		Name:        "Mi Empresa",
		Owner:       domain.User{ID: testOwnerID},
		Address:     domain.Address{Department: "Leon", Municipality: "Leon", AddressLine: "Calle Central"},
		Description: "Una empresa",
		PhoneNumber: "1234-5678",
		Email:       "findme@example.com",
		Website:     "http://web",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	if err := db.Save(context.Background(), savedCompany); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testCompanyID,
			ExpectedErr: nil,
		},
		{
			Name:        "Company Not Found",
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333399"),
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			company, err := db.FindByID(context.Background(), tt.ID)

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

			if company == nil {
				t.Fatal("FindByID() returned nil company")
			}
			if company.ID != tt.ID {
				t.Errorf("FindByID() ID = %v, want %v", company.ID, tt.ID)
			}
			if company.Name != savedCompany.Name {
				t.Errorf("FindByID() Name = %v, want %v", company.Name, savedCompany.Name)
			}
			if company.Owner.ID != testOwnerID {
				t.Errorf("FindByID() Owner.ID = %v, want %v", company.Owner.ID, testOwnerID)
			}
			if company.Email != savedCompany.Email {
				t.Errorf("FindByID() Email = %v, want %v", company.Email, savedCompany.Email)
			}
		})
	}
}

func TestCompanyFindByOwner(t *testing.T) {
	setupCompanyTestData(t)
	db := repository.NewCompanyRepository(TestPool)

	company1 := &domain.Company{
		ID:          testCompanyID,
		Name:        "Empresa Uno",
		Owner:       domain.User{ID: testOwnerID},
		Address:     domain.Address{},
		Description: "Primera empresa",
		PhoneNumber: "1111-1111",
		Email:       "uno@example.com",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	company2 := &domain.Company{
		ID:          testCompanyID2,
		Name:        "Empresa Dos",
		Owner:       domain.User{ID: testOwnerID},
		Address:     domain.Address{},
		Description: "Segunda empresa",
		PhoneNumber: "2222-2222",
		Email:       "dos@example.com",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	if err := db.Save(context.Background(), company1); err != nil {
		t.Fatalf("Save company1: %v", err)
	}
	if err := db.Save(context.Background(), company2); err != nil {
		t.Fatalf("Save company2: %v", err)
	}

	tests := []struct {
		Name         string
		OwnerID      uuid.UUID
		ExpectedLen  int
		ExpectedErr  error
		ExpectedNames []string
	}{
		{
			Name:         "Owner with companies",
			OwnerID:      testOwnerID,
			ExpectedLen:  2,
			ExpectedNames: []string{"Empresa Uno", "Empresa Dos"},
		},
		{
			Name:        "Owner with no companies",
			OwnerID:     uuid.MustParse("11111111-1111-1111-1111-111111111199"),
			ExpectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			companies, err := db.FindByOwner(context.Background(), tt.OwnerID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByOwner() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByOwner() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByOwner() unexpected error: %v", err)
				return
			}

			if len(companies) < tt.ExpectedLen {
				t.Fatalf("FindByOwner() got %d companies, want at least %d", len(companies), tt.ExpectedLen)
			}

			if tt.ExpectedNames != nil {
				found := make(map[string]bool)
				for _, c := range companies {
					found[c.Name] = true
				}
				for _, name := range tt.ExpectedNames {
					if !found[name] {
						t.Errorf("FindByOwner() did not find company '%s'", name)
					}
				}
			}
		})
	}
}

func TestCompanyUpdate(t *testing.T) {
	setupCompanyTestData(t)
	db := repository.NewCompanyRepository(TestPool)

	savedCompany := &domain.Company{
		ID:          testCompanyID,
		Name:        "Empresa Original",
		Owner:       domain.User{ID: testOwnerID},
		Address:     domain.Address{Department: "Leon", Municipality: "Leon", AddressLine: "Calle 1"},
		Description: "Descripcion original",
		PhoneNumber: "1234-5678",
		Email:       "update@example.com",
		Website:     "http://old",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	if err := db.Save(context.Background(), savedCompany); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ExpectedErr error
		update_func func() *domain.Company
	}{
		{
			Name: "Happy path",
			update_func: func() *domain.Company {
				return &domain.Company{
					ID:          testCompanyID,
					Name:        "Empresa Actualizada",
					Owner:       domain.User{ID: testOwnerID},
					Address:     domain.Address{Department: "Leon", Municipality: "Leon", AddressLine: "Calle 1"},
					Description: "Nueva descripcion",
					PhoneNumber: "8765-4321",
					Email:       "update@example.com",
					Website:     "http://new",
					Verified:    true,
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime2,
				}
			},
		},
		{
			Name: "Company not found",
			update_func: func() *domain.Company {
				return &domain.Company{
					ID:          uuid.MustParse("33333333-3333-3333-3333-333333333399"),
					Name:        "Fantasma",
					Owner:       domain.User{ID: testOwnerID},
					Address:     domain.Address{},
					Description: "No existe",
					PhoneNumber: "0000-0000",
					Email:       "ghost@example.com",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
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

	t.Run("Verify updated values", func(t *testing.T) {
		updated, err := db.FindByID(context.Background(), testCompanyID)
		if err != nil {
			t.Fatalf("FindByID() after update error: %v", err)
		}

		if updated.Name != "Empresa Actualizada" {
			t.Errorf("After Update() Name = %v, want 'Empresa Actualizada'", updated.Name)
		}
		if updated.Description != "Nueva descripcion" {
			t.Errorf("After Update() Description = %v, want 'Nueva descripcion'", updated.Description)
		}
		if updated.PhoneNumber != "8765-4321" {
			t.Errorf("After Update() PhoneNumber = %v, want '8765-4321'", updated.PhoneNumber)
		}
	})
}
