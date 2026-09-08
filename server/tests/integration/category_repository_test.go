package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
)

func TestCategoryFindAll(t *testing.T) {
	db := repository.NewCategoryRepository(TestPool)

	tests := []struct {
		Name         string
		SetupCats    []domain.Category
		ExpectedLen  int
		ExpectedErr  error
		ExpectedNames []string
	}{
		{
			Name:        "Empty table",
			SetupCats:   nil,
			ExpectedLen: 0,
		},
		{
			Name: "Single category",
			SetupCats: []domain.Category{
				{ID: uuid.New(), Name: "Tech", Description: "Technology services"},
			},
			ExpectedLen:  1,
			ExpectedNames: []string{"Tech"},
		},
		{
			Name: "Multiple categories",
			SetupCats: []domain.Category{
				{ID: uuid.New(), Name: "Tech", Description: "Technology services"},
				{ID: uuid.New(), Name: "Food", Description: "Food products"},
			},
			ExpectedLen:  2,
			ExpectedNames: []string{"Tech", "Food"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			for i := range tt.SetupCats {
				if err := db.Save(context.Background(), &tt.SetupCats[i]); err != nil {
					t.Fatalf("Save category: %v", err)
				}
			}

			categories, err := db.FindAll(context.Background())

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindAll() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindAll() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindAll() unexpected error: %v", err)
				return
			}

			if len(categories) < tt.ExpectedLen {
				t.Fatalf("FindAll() got %d categories, want at least %d", len(categories), tt.ExpectedLen)
			}

			found := make(map[string]bool)
			for _, cat := range categories {
				found[cat.Name] = true
			}
			for _, name := range tt.ExpectedNames {
				if !found[name] {
					t.Errorf("FindAll() did not find category '%s'", name)
				}
			}
		})
	}
}

func TestCategoryFindByID(t *testing.T) {
	db := repository.NewCategoryRepository(TestPool)

	catID := uuid.New()
	saved := &domain.Category{ID: catID, Name: "Health", Description: "Health and wellness"}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          catID,
			ExpectedErr: nil,
		},
		{
			Name:        "Category Not Found",
			ID:          uuid.New(),
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cat, err := db.FindByID(context.Background(), tt.ID)

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

			if cat == nil {
				t.Fatal("FindByID() returned nil category")
			}
			if cat.ID != tt.ID {
				t.Errorf("FindByID() ID = %v, want %v", cat.ID, tt.ID)
			}
			if cat.Name != saved.Name {
				t.Errorf("FindByID() Name = %v, want %v", cat.Name, saved.Name)
			}
		})
	}
}
