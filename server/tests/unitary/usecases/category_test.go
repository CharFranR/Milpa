package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

func TestCategoryUseCaseGetAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		categories []domain.Category
		repoErr    error
		wantLen    int
		wantErr    error
	}{
		{
			name: "happy path",
			categories: []domain.Category{
				*mustCategory(),
				{ID: testOtherID, Name: "Fruits", Description: "Fresh fruits"},
			},
			wantLen: 2,
		},
		{name: "empty", categories: []domain.Category{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			categoryRepo := newFakeCategoryRepo()
			if tt.repoErr != nil {
				categoryRepo.findAll = func(ctx context.Context) ([]domain.Category, error) {
					return nil, tt.repoErr
				}
			} else {
				categoryRepo.findAll = func(ctx context.Context) ([]domain.Category, error) {
					return tt.categories, nil
				}
			}
			uc := usecases.NewCategoryUseCase(categoryRepo)

			got, err := uc.GetAll(context.Background())

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
				if dto.ID != tt.categories[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.categories[i].ID)
				}
				if dto.Name != tt.categories[i].Name {
					t.Errorf("dto %d name = %q, want %q", i, dto.Name, tt.categories[i].Name)
				}
				if dto.Description != tt.categories[i].Description {
					t.Errorf("dto %d description = %q, want %q", i, dto.Description, tt.categories[i].Description)
				}
			}
		})
	}
}
