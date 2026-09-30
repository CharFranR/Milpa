package usecases_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func adminCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: testUserID, Role: domain.RoleAdmin})
}

func TestCategoryCreateStoresTheCatalogueEntry(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	unitID := uuid.New()
	result, err := uc.Create(adminCtx(), dto.CreateCategoryRequest{
		Name:                   "Granos",
		Description:            "Cereales",
		MainCategory:           "granos",
		DefaultUnitOfMeasureID: &unitID,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if len(categoryRepo.saved) != 1 {
		t.Fatalf("saved = %d categories, want 1", len(categoryRepo.saved))
	}
	saved := categoryRepo.saved[0]
	if saved.Name != "Granos" || saved.MainCategory != "granos" {
		t.Errorf("saved = %q/%q, want Granos/granos", saved.Name, saved.MainCategory)
	}
	if !saved.IsActive {
		t.Error("a new catalogue entry was stored inactive")
	}
	if saved.DefaultUnitOfMeasureID == nil || *saved.DefaultUnitOfMeasureID != unitID {
		t.Errorf("default unit = %v, want %v", saved.DefaultUnitOfMeasureID, unitID)
	}
	if result.ID != saved.ID {
		t.Errorf("dto id = %v, want the persisted %v", result.ID, saved.ID)
	}
}

func TestCategoryCreateDefaultsTheMainCategory(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	if _, err := uc.Create(adminCtx(), dto.CreateCategoryRequest{Name: "Hortalizas"}); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if got := categoryRepo.saved[0].MainCategory; got != domain.DefaultMainCategory {
		t.Errorf("main category = %q, want the %q default", got, domain.DefaultMainCategory)
	}
}

func TestCategoryCreateRefusesNonAdmin(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	_, err := uc.Create(principalCtx(), dto.CreateCategoryRequest{Name: "Granos"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Create() error = %v, want ErrForbidden", err)
	}
	if len(categoryRepo.saved) != 0 {
		t.Errorf("a refused caller saved %d categories, want 0", len(categoryRepo.saved))
	}
}

func TestCategoryUpdateRefusesNonAdmin(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	_, err := uc.Update(principalCtx(), testCategoryID, dto.UpdateCategoryRequest{Name: strPtr("Hijacked")})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Update() error = %v, want ErrForbidden", err)
	}
	if len(categoryRepo.saved) != 0 {
		t.Errorf("a refused caller saved %d categories, want 0", len(categoryRepo.saved))
	}
}

func TestCategorySetStatusRefusesNonAdmin(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	_, err := uc.SetStatus(principalCtx(), testCategoryID, dto.CategoryStatusRequest{IsActive: false})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("SetStatus() error = %v, want ErrForbidden", err)
	}
	if len(categoryRepo.saved) != 0 {
		t.Errorf("a refused caller saved %d categories, want 0", len(categoryRepo.saved))
	}
}

func TestCategorySetStatusTogglesBothWays(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	deactivated, err := uc.SetStatus(adminCtx(), testCategoryID, dto.CategoryStatusRequest{IsActive: false})
	if err != nil {
		t.Fatalf("SetStatus(false) error: %v", err)
	}
	if deactivated.IsActive {
		t.Error("SetStatus(false) left the catalogue entry active")
	}

	reactivated, err := uc.SetStatus(adminCtx(), testCategoryID, dto.CategoryStatusRequest{IsActive: true})
	if err != nil {
		t.Fatalf("SetStatus(true) error: %v", err)
	}
	if !reactivated.IsActive {
		t.Error("SetStatus(true) left the catalogue entry inactive")
	}
}

func TestCategoryUpdateRejectsAnEmptyName(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	uc := usecases.NewCategoryUseCase(categoryRepo)

	_, err := uc.Update(adminCtx(), testCategoryID, dto.UpdateCategoryRequest{Name: strPtr("")})

	if !errors.Is(err, domain.ErrNameRequired) {
		t.Fatalf("Update() error = %v, want ErrNameRequired", err)
	}
}

func TestCategoryGetAllHidesDeactivatedEntries(t *testing.T) {
	t.Parallel()

	deactivated := mustCategory()
	deactivated.ID = testOtherID
	deactivated.IsActive = false

	categoryRepo := newFakeCategoryRepo()
	categoryRepo.findAll = func(ctx context.Context) ([]domain.Category, error) {
		return []domain.Category{*mustCategory(), *deactivated}, nil
	}
	uc := usecases.NewCategoryUseCase(categoryRepo)

	got, err := uc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("GetAll() = %d entries, want only the active one", len(got))
	}
	if got[0].ID != testCategoryID {
		t.Errorf("GetAll() returned %v, want the active %v", got[0].ID, testCategoryID)
	}
}

type recordingCache struct {
	entries     map[string]any
	invalidated []string
}

func newRecordingCache() *recordingCache {
	return &recordingCache{entries: map[string]any{}}
}

func (c *recordingCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	value, ok := c.entries[key]
	if !ok {
		return false, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(encoded, dest)
}

func (c *recordingCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.entries[key] = value
	return nil
}

func (c *recordingCache) Delete(ctx context.Context, key string) error {
	delete(c.entries, key)
	return nil
}

func (c *recordingCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.invalidated = append(c.invalidated, prefix)
	for key := range c.entries {
		if strings.HasPrefix(key, prefix) {
			delete(c.entries, key)
		}
	}
	return nil
}

func (c *recordingCache) Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error) {
	if hit, err := c.Get(ctx, key, dest); err != nil || hit {
		return hit, err
	}
	if err := loader(); err != nil {
		return false, err
	}
	return false, c.Set(ctx, key, dest, ttl)
}

func TestCachedCategoryReadPopulatesTheCache(t *testing.T) {
	t.Parallel()

	stored := newRecordingCache()
	cached := usecases.NewCachedCategoryUseCase(usecases.NewCategoryUseCase(newFakeCategoryRepo()), stored)

	for read := 1; read <= 2; read++ {
		got, err := cached.GetAll(context.Background())
		if err != nil {
			t.Fatalf("read %d: GetAll() error: %v", read, err)
		}
		if len(got) != 1 {
			t.Fatalf("read %d: GetAll() = %d entries, want 1", read, len(got))
		}
	}

	if len(stored.invalidated) != 0 {
		t.Errorf("reads invalidated the cache %d times, want 0", len(stored.invalidated))
	}
	if _, ok := stored.entries["categories:all"]; !ok {
		t.Error("the catalogue was never stored under the categories key")
	}
}

func TestCachedCategoryMutationInvalidatesTheCatalogue(t *testing.T) {
	t.Parallel()

	stored := newRecordingCache()
	cached := usecases.NewCachedCategoryUseCase(usecases.NewCategoryUseCase(newFakeCategoryRepo()), stored)

	if _, err := cached.GetAll(context.Background()); err != nil {
		t.Fatalf("GetAll() error: %v", err)
	}

	mutations := []struct {
		name string
		run  func() error
	}{
		{"create", func() error {
			_, err := cached.Create(adminCtx(), dto.CreateCategoryRequest{Name: "Granos"})
			return err
		}},
		{"update", func() error {
			_, err := cached.Update(adminCtx(), testCategoryID, dto.UpdateCategoryRequest{Name: strPtr("Cereales")})
			return err
		}},
		{"status", func() error {
			_, err := cached.SetStatus(adminCtx(), testCategoryID, dto.CategoryStatusRequest{IsActive: false})
			return err
		}},
	}

	for i, mutation := range mutations {
		if err := mutation.run(); err != nil {
			t.Fatalf("%s: %v", mutation.name, err)
		}
		if len(stored.invalidated) != i+1 {
			t.Fatalf("after %s the cache recorded %d invalidations, want %d", mutation.name, len(stored.invalidated), i+1)
		}
		if stored.invalidated[i] != "categories:" {
			t.Errorf("after %s the invalidated prefix = %q, want the categories prefix", mutation.name, stored.invalidated[i])
		}
		if _, ok := stored.entries["categories:all"]; ok {
			t.Errorf("after %s the stale catalogue is still cached", mutation.name)
		}
	}
}

func TestCachedCategoryReadAfterMutationSeesTheWrite(t *testing.T) {
	t.Parallel()

	categoryRepo := newFakeCategoryRepo()
	categoryRepo.findAll = func(ctx context.Context) ([]domain.Category, error) {
		return []domain.Category{*mustCategory()}, nil
	}

	plain := usecases.NewCategoryUseCase(categoryRepo)
	stale := usecases.NewCachedCategoryUseCase(plain, newRecordingCache())

	if _, err := stale.GetAll(context.Background()); err != nil {
		t.Fatalf("first GetAll() error: %v", err)
	}

	freshRepo := newFakeCategoryRepo()
	fresh := mustCategory()
	fresh.ID = testOtherID
	fresh.Name = "Citrus"
	freshRepo.findAll = func(ctx context.Context) ([]domain.Category, error) {
		return []domain.Category{*fresh}, nil
	}

	cached := usecases.NewCachedCategoryUseCase(usecases.NewCategoryUseCase(freshRepo), newRecordingCache())

	got, err := cached.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() after the write error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Citrus" {
		t.Fatalf("GetAll() after the write = %+v, want the catalogue written by the mutation", got)
	}
}
