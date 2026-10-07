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

func (r *fakeUnitOfMeasureRepo) FindAll(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return r.units, nil
}

func (r *fakeUnitOfMeasureRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	for i := range r.units {
		if r.units[i].ID == id {
			copied := r.units[i]
			return &copied, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeUnitOfMeasureRepo) Save(ctx context.Context, unit *domain.UnitOfMeasure) error {
	return nil
}

type adminUnitRepo struct {
	list    []domain.UnitOfMeasure
	findAll []domain.UnitOfMeasure
	unit    *domain.UnitOfMeasure
	saveErr error
	saved   []*domain.UnitOfMeasure
}

func (r *adminUnitRepo) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return r.list, nil
}

func (r *adminUnitRepo) FindAll(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return r.findAll, nil
}

func (r *adminUnitRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	if r.unit == nil {
		return nil, domain.ErrNotFound
	}
	copied := *r.unit
	copied.ID = id
	return &copied, nil
}

func (r *adminUnitRepo) Save(ctx context.Context, unit *domain.UnitOfMeasure) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	copied := *unit
	r.saved = append(r.saved, &copied)
	return nil
}

func adminUnitContext() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: testUserID, Role: domain.RoleAdmin})
}

func TestUnitOfMeasureAdminWritesRefuseANonAdmin(t *testing.T) {
	t.Parallel()

	verbs := []struct {
		name string
		call func(uc *usecases.UnitOfMeasureUseCaseImpl) error
	}{
		{
			name: "create",
			call: func(uc *usecases.UnitOfMeasureUseCaseImpl) error {
				_, err := uc.Create(principalCtx(), dto.CreateUnitOfMeasureRequest{Code: "kg", Name: "Kilogramo"})
				return err
			},
		},
		{
			name: "update",
			call: func(uc *usecases.UnitOfMeasureUseCaseImpl) error {
				_, err := uc.Update(principalCtx(), testUnitOfMeasureID, dto.UpdateUnitOfMeasureRequest{Name: strPtr("Hijacked")})
				return err
			},
		},
		{
			name: "status",
			call: func(uc *usecases.UnitOfMeasureUseCaseImpl) error {
				_, err := uc.SetStatus(principalCtx(), testUnitOfMeasureID, dto.UnitOfMeasureStatusRequest{IsActive: false})
				return err
			},
		},
		{
			name: "list all",
			call: func(uc *usecases.UnitOfMeasureUseCaseImpl) error {
				_, err := uc.ListAll(principalCtx())
				return err
			},
		},
	}

	for _, verb := range verbs {
		t.Run(verb.name, func(t *testing.T) {
			t.Parallel()

			repo := &adminUnitRepo{unit: &domain.UnitOfMeasure{ID: testUnitOfMeasureID, Code: "kg", Name: "Kilogramo", IsActive: true}}
			uc := usecases.NewUnitOfMeasureUseCase(repo)

			err := verb.call(uc)
			if !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("error = %v, want ErrForbidden", err)
			}
			if len(repo.saved) != 0 {
				t.Errorf("a refused caller saved %d units, want 0", len(repo.saved))
			}
		})
	}
}

func TestUnitOfMeasureAdminWritesRequireAuthentication(t *testing.T) {
	t.Parallel()

	repo := &adminUnitRepo{unit: &domain.UnitOfMeasure{ID: testUnitOfMeasureID, Code: "kg", Name: "Kilogramo", IsActive: true}}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	if _, err := uc.Create(context.Background(), dto.CreateUnitOfMeasureRequest{Code: "kg", Name: "Kilogramo"}); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("Create() error = %v, want ErrUnauthenticated", err)
	}
	if _, err := uc.Update(context.Background(), testUnitOfMeasureID, dto.UpdateUnitOfMeasureRequest{Name: strPtr("Kilogramo")}); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("Update() error = %v, want ErrUnauthenticated", err)
	}
	if _, err := uc.SetStatus(context.Background(), testUnitOfMeasureID, dto.UnitOfMeasureStatusRequest{IsActive: false}); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("SetStatus() error = %v, want ErrUnauthenticated", err)
	}
	if _, err := uc.ListAll(context.Background()); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("ListAll() error = %v, want ErrUnauthenticated", err)
	}
}

func TestUnitOfMeasureAdminCreateNormalizesTheInput(t *testing.T) {
	t.Parallel()

	repo := &adminUnitRepo{}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	result, err := uc.Create(adminUnitContext(), dto.CreateUnitOfMeasureRequest{Code: "  KG  ", Name: "  Kilogramo  "})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if result.Code != "kg" || result.Name != "Kilogramo" {
		t.Errorf("dto = %+v, want the trimmed lowercase code and trimmed name", result)
	}
	if !result.IsActive {
		t.Error("Create() returned an inactive unit")
	}
	if len(repo.saved) != 1 || repo.saved[0].Code != "kg" {
		t.Fatalf("saved = %+v, want the normalized unit", repo.saved)
	}
}

func TestUnitOfMeasureAdminUpdateLeavesUntouchedFields(t *testing.T) {
	t.Parallel()

	repo := &adminUnitRepo{unit: &domain.UnitOfMeasure{ID: testUnitOfMeasureID, Code: "kg", Name: "Kilogramo", IsActive: true}}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	result, err := uc.Update(adminUnitContext(), testUnitOfMeasureID, dto.UpdateUnitOfMeasureRequest{Name: strPtr("  Kilo  ")})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if result.Code != "kg" {
		t.Errorf("code = %q, want the untouched kg", result.Code)
	}
	if result.Name != "Kilo" {
		t.Errorf("name = %q, want the trimmed Kilo", result.Name)
	}
	if len(repo.saved) != 1 || repo.saved[0].Code != "kg" || repo.saved[0].Name != "Kilo" {
		t.Fatalf("saved = %+v, want only the name changed", repo.saved)
	}
}

func TestUnitOfMeasureAdminSetStatusTogglesBothWays(t *testing.T) {
	t.Parallel()

	repo := &adminUnitRepo{unit: &domain.UnitOfMeasure{ID: testUnitOfMeasureID, Code: "kg", Name: "Kilogramo", IsActive: true}}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	deactivated, err := uc.SetStatus(adminUnitContext(), testUnitOfMeasureID, dto.UnitOfMeasureStatusRequest{IsActive: false})
	if err != nil {
		t.Fatalf("SetStatus(false) error: %v", err)
	}
	if deactivated.IsActive {
		t.Error("SetStatus(false) left the unit active")
	}

	reactivated, err := uc.SetStatus(adminUnitContext(), testUnitOfMeasureID, dto.UnitOfMeasureStatusRequest{IsActive: true})
	if err != nil {
		t.Fatalf("SetStatus(true) error: %v", err)
	}
	if !reactivated.IsActive {
		t.Error("SetStatus(true) left the unit inactive")
	}
}

func TestUnitOfMeasureAdminCreateSurfacesADuplicate(t *testing.T) {
	t.Parallel()

	repo := &adminUnitRepo{saveErr: domain.ErrDuplicate}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	_, err := uc.Create(adminUnitContext(), dto.CreateUnitOfMeasureRequest{Code: "kg", Name: "Kilogramo"})
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("Create() error = %v, want ErrDuplicate", err)
	}
}

func TestUnitOfMeasureAdminListAllKeepsInactiveRows(t *testing.T) {
	t.Parallel()

	active := domain.UnitOfMeasure{ID: testUnitOfMeasureID, Code: "kg", Name: "Kilogramo", IsActive: true}
	retired := domain.UnitOfMeasure{ID: testOtherID, Code: "qq", Name: "Quintal", IsActive: false}

	repo := &adminUnitRepo{
		list:    []domain.UnitOfMeasure{active},
		findAll: []domain.UnitOfMeasure{active, retired},
	}
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	all, err := uc.ListAll(adminUnitContext())
	if err != nil {
		t.Fatalf("ListAll() error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("ListAll() = %d units, want the active and the inactive one", len(all))
	}

	public, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(public) != 1 {
		t.Fatalf("List() = %d units, want only the active one", len(public))
	}
	if public[0].ID != testUnitOfMeasureID {
		t.Errorf("List() returned %v, want the active unit", public[0].ID)
	}
}
