package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

type fakeUnitOfMeasureRepo struct {
	units   []domain.UnitOfMeasure
	listErr error
}

func (r *fakeUnitOfMeasureRepo) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.units, nil
}

func TestUnitOfMeasureUseCaseList(t *testing.T) {
	t.Parallel()

	kilo := uuid.MustParse("c2dba60d-a701-4e93-990a-f2fb814ad114")
	quintal := uuid.MustParse("d5862954-4d0f-4ef6-9f4b-4df5bf8229aa")

	t.Run("maps every unit", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUnitOfMeasureRepo{units: []domain.UnitOfMeasure{
			{ID: kilo, Code: "kg", Name: "Kilogramo", IsActive: true},
			{ID: quintal, Code: "qq", Name: "Quintal", IsActive: true},
		}}
		uc := usecases.NewUnitOfMeasureUseCase(repo)

		got, err := uc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("dtos = %d, want 2", len(got))
		}
		if got[0].ID != kilo || got[0].Code != "kg" || got[0].Name != "Kilogramo" {
			t.Errorf("dto 0 = %+v, want kg/Kilogramo", got[0])
		}
		if got[1].Code != "qq" || got[1].Name != "Quintal" {
			t.Errorf("dto 1 = %+v, want qq/Quintal", got[1])
		}
	})

	t.Run("empty catalogue stays empty instead of nil", func(t *testing.T) {
		t.Parallel()

		uc := usecases.NewUnitOfMeasureUseCase(&fakeUnitOfMeasureRepo{})

		got, err := uc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("dtos = nil, want empty slice so the JSON envelope serializes as []")
		}
		if len(got) != 0 {
			t.Fatalf("dtos = %d, want 0", len(got))
		}
	})

	t.Run("repository error propagates", func(t *testing.T) {
		t.Parallel()

		uc := usecases.NewUnitOfMeasureUseCase(&fakeUnitOfMeasureRepo{listErr: errors.New("boom")})

		if _, err := uc.List(context.Background()); err == nil {
			t.Fatal("err = nil, want the repository failure")
		}
	})
}