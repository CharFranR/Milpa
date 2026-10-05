package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/database"
	"milpa/internal/auth"
)

func catalogueAdminContext() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{
		UserID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
		Role:   domain.RoleAdmin,
	})
}

func catalogueFarmerContext() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{
		UserID: uuid.MustParse("88888888-8888-8888-8888-888888888888"),
		Role:   domain.RoleAgricultor,
	})
}

func TestSeededCatalogueIsProducedByTheMigrationChain(t *testing.T) {
	ctx := context.Background()

	src, err := database.NewMigrationSource()
	if err != nil {
		t.Fatalf("NewMigrationSource() error: %v", err)
	}

	scratchURL := scratchDatabaseURL(t, "catalogue_seed_probe")

	m, err := migrate.NewWithSourceInstance("iofs", src, scratchURL)
	if err != nil {
		t.Fatalf("build the migration runner: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("m.Up() error: %v", err)
	}

	probe, err := pgxpool.New(ctx, scratchURL)
	if err != nil {
		t.Fatalf("connect to the scratch database: %v", err)
	}
	defer probe.Close()

	rows, err := probe.Query(ctx, `SELECT code, name FROM units_of_measure ORDER BY code`)
	if err != nil {
		t.Fatalf("read units_of_measure: %v", err)
	}
	defer rows.Close()

	units := map[string]string{}
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			t.Fatalf("scan unit: %v", err)
		}
		units[code] = name
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read units_of_measure: %v", err)
	}

	for _, code := range []string{"kg", "qq", "lb", "tn", "unidad", "docena"} {
		if _, ok := units[code]; !ok {
			t.Errorf("the seeded unit %q is missing", code)
		}
	}
	if len(units) != 6 {
		t.Errorf("seeded units = %d, want 6", len(units))
	}

	rows, err = probe.Query(ctx, `SELECT name, main_category, default_unit_of_measure_id FROM categories ORDER BY name`)
	if err != nil {
		t.Fatalf("read categories: %v", err)
	}
	defer rows.Close()

	categories := map[string]struct {
		mainCategory string
		unitID       *uuid.UUID
	}{}
	for rows.Next() {
		var name string
		entry := struct {
			mainCategory string
			unitID       *uuid.UUID
		}{}
		if err := rows.Scan(&name, &entry.mainCategory, &entry.unitID); err != nil {
			t.Fatalf("scan category: %v", err)
		}
		categories[name] = entry
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read categories: %v", err)
	}

	for name, want := range map[string]string{"Frutales": "frutales", "Cítricos": "citrusos"} {
		entry, ok := categories[name]
		if !ok {
			t.Errorf("the seeded category %q is missing", name)
			continue
		}
		if entry.mainCategory != want {
			t.Errorf("category %q main_category = %q, want %q", name, entry.mainCategory, want)
		}
		if entry.unitID == nil {
			t.Errorf("category %q has no default unit of measure", name)
		}
	}
}

func TestAdminCatalogueLifecycleAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	uc := usecases.NewCategoryUseCase(repository.NewCategoryRepository(TestPool))
	admin := catalogueAdminContext()

	unit := seededUnitOfMeasure(t, "qq")

	created, err := uc.Create(admin, dto.CreateCategoryRequest{
		Name:                   "Granos",
		Description:            "Cereales",
		MainCategory:           "granos",
		DefaultUnitOfMeasureID: &unit,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if !created.IsActive {
		t.Error("Create() returned an inactive entry")
	}

	visible, err := uc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error: %v", err)
	}
	if !containsCategory(visible, created.ID) {
		t.Fatalf("the public catalogue does not carry the created entry %s", created.ID)
	}

	updated, err := uc.Update(admin, created.ID, dto.UpdateCategoryRequest{
		Description:  catalogueStringPtr("Cereales y legumbres"),
		MainCategory: catalogueStringPtr("granos"),
	})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if updated.Description != "Cereales y legumbres" {
		t.Errorf("Update() description = %q, want the requested one", updated.Description)
	}

	reloaded, err := uc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() after the update error: %v", err)
	}
	for _, entry := range reloaded {
		if entry.ID == created.ID && entry.Description != "Cereales y legumbres" {
			t.Errorf("the public catalogue still reads %q, want the updated description", entry.Description)
		}
	}

	deactivated, err := uc.SetStatus(admin, created.ID, dto.CategoryStatusRequest{IsActive: false})
	if err != nil {
		t.Fatalf("SetStatus() error: %v", err)
	}
	if deactivated.IsActive {
		t.Fatal("SetStatus(false) left the entry active")
	}

	visible, err = uc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() after the deactivation error: %v", err)
	}
	if containsCategory(visible, created.ID) {
		t.Errorf("the public catalogue still returns the deactivated entry %s", created.ID)
	}

	stored, err := repository.NewCategoryRepository(TestPool).FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if stored.IsActive {
		t.Error("the deactivated entry is still active in storage")
	}
	if stored.DefaultUnitOfMeasureID == nil || *stored.DefaultUnitOfMeasureID != unit {
		t.Errorf("stored default unit = %v, want %v", stored.DefaultUnitOfMeasureID, unit)
	}
}

func TestAdminCatalogueWritesRefuseANonAdminAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	uc := usecases.NewCategoryUseCase(repository.NewCategoryRepository(TestPool))
	farmer := catalogueFarmerContext()

	if _, err := uc.Create(farmer, dto.CreateCategoryRequest{Name: "Granos"}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Create() error = %v, want ErrForbidden", err)
	}

	category, err := domain.NewCategory("Granos")
	if err != nil {
		t.Fatalf("NewCategory() error: %v", err)
	}
	if err := repository.NewCategoryRepository(TestPool).Save(context.Background(), category); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := uc.Update(farmer, category.ID, dto.UpdateCategoryRequest{Name: catalogueStringPtr("Hijacked")}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Update() error = %v, want ErrForbidden", err)
	}
	if _, err := uc.SetStatus(farmer, category.ID, dto.CategoryStatusRequest{IsActive: false}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("SetStatus() error = %v, want ErrForbidden", err)
	}

	stored, err := repository.NewCategoryRepository(TestPool).FindByID(context.Background(), category.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if stored.Name != "Granos" || !stored.IsActive {
		t.Errorf("a refused caller changed the entry to %+v", stored)
	}
}

func TestCategorySaveRefusesADuplicateName(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewCategoryRepository(TestPool)

	first, err := domain.NewCategory("Granos")
	if err != nil {
		t.Fatalf("NewCategory() error: %v", err)
	}
	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	second, err := domain.NewCategory("Granos")
	if err != nil {
		t.Fatalf("NewCategory() error: %v", err)
	}
	if err := repo.Save(context.Background(), second); !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("Save() of a duplicate name = %v, want ErrDuplicate", err)
	}
}

func containsCategory(categories []*dto.CategoryDTO, id uuid.UUID) bool {
	for _, category := range categories {
		if category.ID == id {
			return true
		}
	}
	return false
}

func catalogueStringPtr(s string) *string {
	return &s
}

func seededUnitOfMeasure(t *testing.T, code string) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	if err := TestPool.QueryRow(context.Background(),
		`SELECT id FROM units_of_measure WHERE code = $1`, code).Scan(&id); err != nil {
		t.Fatalf("read the seeded unit %q: %v", code, err)
	}
	return id
}
