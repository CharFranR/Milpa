package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

func TestUnitOfMeasureSaveAndFindByIDAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewUnitOfMeasureRepository(TestPool)
	ctx := context.Background()

	unit, err := domain.NewUnitOfMeasure("  T1  ", "  Tonelada métrica  ")
	if err != nil {
		t.Fatalf("NewUnitOfMeasure() error: %v", err)
	}
	cleanupUnit(t, unit.ID)
	if err := repo.Save(ctx, unit); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	stored, err := repo.FindByID(ctx, unit.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if stored.Code != "t1" || stored.Name != "Tonelada métrica" || !stored.IsActive {
		t.Errorf("stored = %+v, want the normalized t1/Tonelada métrica active unit", stored)
	}
}

func TestUnitOfMeasureSaveUpsertsAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewUnitOfMeasureRepository(TestPool)
	ctx := context.Background()

	unit, err := domain.NewUnitOfMeasure("t2", "Tonelada")
	if err != nil {
		t.Fatalf("NewUnitOfMeasure() error: %v", err)
	}
	cleanupUnit(t, unit.ID)
	if err := repo.Save(ctx, unit); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := unit.SetCode("t2b"); err != nil {
		t.Fatalf("SetCode() error: %v", err)
	}
	if err := unit.SetName("Tonelada corta"); err != nil {
		t.Fatalf("SetName() error: %v", err)
	}
	unit.Deactivate()

	if err := repo.Save(ctx, unit); err != nil {
		t.Fatalf("Save() upsert error: %v", err)
	}

	stored, err := repo.FindByID(ctx, unit.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if stored.Code != "t2b" || stored.Name != "Tonelada corta" {
		t.Errorf("stored = %+v, want the upserted code and name", stored)
	}
	if stored.IsActive {
		t.Error("the upsert did not persist is_active = false")
	}
}

func TestUnitOfMeasureFindByIDMissingAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewUnitOfMeasureRepository(TestPool)

	if _, err := repo.FindByID(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want ErrNotFound", err)
	}
}

func TestUnitOfMeasureSaveRefusesADuplicateCodeAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewUnitOfMeasureRepository(TestPool)
	ctx := context.Background()

	first, err := domain.NewUnitOfMeasure("t3", "Tonelada")
	if err != nil {
		t.Fatalf("NewUnitOfMeasure() error: %v", err)
	}
	cleanupUnit(t, first.ID)
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	second, err := domain.NewUnitOfMeasure("t3", "Tonelada duplicada")
	if err != nil {
		t.Fatalf("NewUnitOfMeasure() error: %v", err)
	}
	if err := repo.Save(ctx, second); !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("Save() of a duplicate code = %v, want ErrDuplicate", err)
	}
}

func TestUnitOfMeasureFindAllKeepsInactiveRowsWhileListHidesThemAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewUnitOfMeasureRepository(TestPool)
	ctx := context.Background()

	unit, err := domain.NewUnitOfMeasure("t4", "Tonelada")
	if err != nil {
		t.Fatalf("NewUnitOfMeasure() error: %v", err)
	}
	cleanupUnit(t, unit.ID)
	if err := repo.Save(ctx, unit); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error: %v", err)
	}
	if !containsUnit(all, unit.ID) {
		t.Fatalf("FindAll() does not carry the saved unit %s", unit.ID)
	}

	visible, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !containsUnit(visible, unit.ID) {
		t.Fatalf("List() dropped the active unit %s", unit.ID)
	}

	unit.Deactivate()
	if err := repo.Save(ctx, unit); err != nil {
		t.Fatalf("Save() after the deactivation error: %v", err)
	}

	all, err = repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() after the deactivation error: %v", err)
	}
	if !containsUnit(all, unit.ID) {
		t.Errorf("FindAll() dropped the deactivated unit %s", unit.ID)
	}
	for _, entry := range all {
		if entry.ID == unit.ID && entry.IsActive {
			t.Errorf("FindAll() reports the deactivated unit %s as active", unit.ID)
		}
	}

	visible, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List() after the deactivation error: %v", err)
	}
	if containsUnit(visible, unit.ID) {
		t.Errorf("List() still returns the deactivated unit %s", unit.ID)
	}
}

func containsUnit(units []domain.UnitOfMeasure, id uuid.UUID) bool {
	for _, unit := range units {
		if unit.ID == id {
			return true
		}
	}
	return false
}

func cleanupUnit(t *testing.T, id uuid.UUID) {
	t.Helper()

	t.Cleanup(func() {
		if _, err := TestPool.Exec(context.Background(), "DELETE FROM units_of_measure WHERE id = $1", id); err != nil {
			t.Errorf("delete unit %s: %v", id, err)
		}
	})
}
