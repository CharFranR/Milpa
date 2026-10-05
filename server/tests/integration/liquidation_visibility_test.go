package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/secondary/repository"
	"milpa/internal/auth"
)

var (
	liqSupplierID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	liqOtherSupID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	liqLocationID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000003")
	liqCreatedAt  = time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
)

// seedLiquidationUser creates a user a liquidation can reference.
func seedLiquidationUser(t *testing.T, id uuid.UUID) {
	t.Helper()
	seedLiquidationUserWithRole(t, id, 2)
}

func seedLiquidationUserWithRole(t *testing.T, id uuid.UUID, role int) {
	t.Helper()

	if _, err := TestPool.Exec(context.Background(),
		`INSERT INTO users (id, first_name, last_name, role, email, phone_number, password_hash, created_at, updated_at)
		 VALUES ($1, 'Supplier', 'Test', $2, $3, '555-0000', 'hash', $4, $4)`,
		id, role, id.String()+"@milpa.com.ni", liqCreatedAt,
	); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

func anonymousViewer() port.LiquidationViewer {
	return port.LiquidationViewer{}
}

func viewerOf(id uuid.UUID) port.LiquidationViewer {
	return port.LiquidationViewer{ID: id}
}

func mayoristaViewerOf(id uuid.UUID) port.LiquidationViewer {
	return port.LiquidationViewer{ID: id, Role: domain.RoleCompradorMayoristaDetallista}
}

func corporateViewerOf(id uuid.UUID) port.LiquidationViewer {
	return port.LiquidationViewer{ID: id, Role: domain.RoleCompradorMayoristaCorporativo}
}

func seedLiquidation(t *testing.T, supplierID uuid.UUID, name, visibility string) *domain.Liquidation {
	t.Helper()

	liq, err := domain.NewLiquidation(supplierID, name, 10, "kg", 100, 10, liqCreatedAt)
	if err != nil {
		t.Fatalf("NewLiquidation: %v", err)
	}
	liq.LocationID = liqLocationID
	if visibility != "" {
		if err := liq.UpdateVisibility(visibility, liqCreatedAt); err != nil {
			t.Fatalf("UpdateVisibility(%q): %v", visibility, err)
		}
	}

	repo := repository.NewLiquidationRepository(TestPool)
	if err := repo.Save(context.Background(), liq); err != nil {
		t.Fatalf("Save liquidation: %v", err)
	}

	return liq
}

func idsOf(liquidations []domain.Liquidation) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(liquidations))
	for i := range liquidations {
		ids = append(ids, liquidations[i].ID)
	}
	return ids
}

func contains(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// TestFindOpenHonoursVisibility is the defect: FindOpen was
// "WHERE status = 'open'" with no visibility predicate, and the route is
// unauthenticated, so a liquidation its supplier marked wholesale was listed to
// every anonymous caller.
func TestFindOpenHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	repo := repository.NewLiquidationRepository(TestPool)

	t.Run("anonymous sees only public", func(t *testing.T) {
		open, err := repo.FindOpen(ctx, anonymousViewer())
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		ids := idsOf(open)
		if !contains(ids, public.ID) {
			t.Errorf("open = %v, want the public liquidation %s", ids, public.ID)
		}
		if contains(ids, restricted.ID) {
			t.Errorf("open = %v, must not contain the private liquidation %s", ids, restricted.ID)
		}
	})

	t.Run("owner sees their own regardless of visibility", func(t *testing.T) {
		open, err := repo.FindOpen(ctx, viewerOf(liqSupplierID))
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		ids := idsOf(open)
		if !contains(ids, restricted.ID) {
			t.Errorf("open = %v, want the owner's private liquidation %s", ids, restricted.ID)
		}
		if !contains(ids, public.ID) {
			t.Errorf("open = %v, want the public liquidation %s", ids, public.ID)
		}
	})

	t.Run("a mayorista sees the restricted one", func(t *testing.T) {
		mayoristaID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000004")
		seedLiquidationUserWithRole(t, mayoristaID, 3)

		open, err := repo.FindOpen(ctx, mayoristaViewerOf(mayoristaID))
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		if !contains(idsOf(open), restricted.ID) {
			t.Errorf("open = %v, want the restricted liquidation %s for a mayorista", idsOf(open), restricted.ID)
		}
	})

	t.Run("each mayorista level sees its own level only", func(t *testing.T) {
		retailID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000005")
		corporateID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000006")
		seedLiquidationUserWithRole(t, retailID, 3)
		seedLiquidationUserWithRole(t, corporateID, 4)

		retailOnly := seedLiquidation(t, liqSupplierID, "Retail-only corn", "wholesale_retail")
		corporateOnly := seedLiquidation(t, liqSupplierID, "Corporate-only corn", "wholesale_corporate")

		retailOpen, err := repo.FindOpen(ctx, mayoristaViewerOf(retailID))
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		if !contains(idsOf(retailOpen), retailOnly.ID) {
			t.Errorf("a mayorista detallista is missing its wholesale_retail row: %v", idsOf(retailOpen))
		}
		if contains(idsOf(retailOpen), corporateOnly.ID) {
			t.Errorf("a mayorista detallista must not see wholesale_corporate: %v", idsOf(retailOpen))
		}

		corporateOpen, err := repo.FindOpen(ctx, corporateViewerOf(corporateID))
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		if !contains(idsOf(corporateOpen), corporateOnly.ID) {
			t.Errorf("a mayorista corporativo is missing its wholesale_corporate row: %v", idsOf(corporateOpen))
		}
		if contains(idsOf(corporateOpen), retailOnly.ID) {
			t.Errorf("a mayorista corporativo must not see wholesale_retail: %v", idsOf(corporateOpen))
		}
	})

	t.Run("a different authenticated user sees only public", func(t *testing.T) {
		seedLiquidationUser(t, liqOtherSupID)

		open, err := repo.FindOpen(ctx, viewerOf(liqOtherSupID))
		if err != nil {
			t.Fatalf("FindOpen: %v", err)
		}
		ids := idsOf(open)
		if contains(ids, restricted.ID) {
			t.Errorf("open = %v, must not contain another supplier's private liquidation", ids)
		}
		if !contains(ids, public.ID) {
			t.Errorf("open = %v, want the public liquidation %s", ids, public.ID)
		}
	})
}

// TestFindVisibleByIDHonoursVisibility is the same defect on the single read:
// FindByID had no visibility predicate either.
func TestFindVisibleByIDHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	repo := repository.NewLiquidationRepository(TestPool)

	tests := []struct {
		name    string
		id      uuid.UUID
		viewer  port.LiquidationViewer
		wantErr error
	}{
		{name: "anonymous reads a public liquidation", id: public.ID, viewer: anonymousViewer()},
		{name: "anonymous cannot read a private one", id: restricted.ID, viewer: anonymousViewer(), wantErr: domain.ErrNotFound},
		{name: "owner reads their own private one", id: restricted.ID, viewer: viewerOf(liqSupplierID)},
		{name: "another user cannot read it", id: restricted.ID, viewer: viewerOf(liqOtherSupID), wantErr: domain.ErrNotFound},
		{name: "a mayorista reads a restricted one", id: restricted.ID, viewer: mayoristaViewerOf(uuid.New())},
		{name: "unknown id", id: uuid.New(), viewer: viewerOf(liqSupplierID), wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindVisibleByID(ctx, tt.id, tt.viewer)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got the liquidation %+v", tt.wantErr, got)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("id = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestFindBySupplierHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	repo := repository.NewLiquidationRepository(TestPool)

	anon, err := repo.FindBySupplier(ctx, liqSupplierID, anonymousViewer())
	if err != nil {
		t.Fatalf("FindBySupplier: %v", err)
	}
	if contains(idsOf(anon), restricted.ID) {
		t.Error("the anonymous supplier listing contains the private liquidation")
	}
	if !contains(idsOf(anon), public.ID) {
		t.Error("the anonymous supplier listing is missing the public liquidation")
	}

	owner, err := repo.FindBySupplier(ctx, liqSupplierID, viewerOf(liqSupplierID))
	if err != nil {
		t.Fatalf("FindBySupplier: %v", err)
	}
	if !contains(idsOf(owner), restricted.ID) {
		t.Error("the owner listing is missing their own private liquidation")
	}

	other, err := repo.FindBySupplier(ctx, liqSupplierID, viewerOf(liqOtherSupID))
	if err != nil {
		t.Fatalf("FindBySupplier: %v", err)
	}
	if contains(idsOf(other), restricted.ID) {
		t.Error("another user's supplier listing contains the private liquidation")
	}
}

// TestFindByIDStaysUnfilteredForTheAuthorisationPaths documents the split: the
// update and delete use cases re-check ownership themselves, and they need to
// see a wholesale liquidation in order to answer forbidden rather than not found.
func TestFindByIDStaysUnfilteredForTheAuthorisationPaths(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	repo := repository.NewLiquidationRepository(TestPool)

	got, err := repo.FindByID(ctx, restricted.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != restricted.ID {
		t.Errorf("id = %v, want %v", got.ID, restricted.ID)
	}
	if got.Visibility != "wholesale" {
		t.Errorf("visibility = %q, want %q", got.Visibility, "wholesale")
	}
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return liqCreatedAt }

// --- use case level ---

func liquidationUC() *usecases.LiquidationUseCaseImpl {
	userRepo := repository.NewUserRepository(TestPool)
	return usecases.NewLiquidationUseCase(repository.NewLiquidationRepository(TestPool), userRepo, fixedClock{})
}

func ownerCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: liqSupplierID, Role: domain.RoleAgricultor})
}

func otherCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: liqOtherSupID, Role: domain.RoleCompradorMinorista})
}

func dtoIDs(dtos []*dto.LiquidationDTO) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(dtos))
	for _, d := range dtos {
		ids = append(ids, d.ID)
	}
	return ids
}

// TestLiquidationUseCaseGetOpenHonoursVisibility is the acceptance gate for the
// listing, exercised through the use case the handler calls.
func TestLiquidationUseCaseGetOpenHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	uc := liquidationUC()

	anon, err := uc.GetOpen(ctx)
	if err != nil {
		t.Fatalf("GetOpen: %v", err)
	}
	if contains(dtoIDs(anon), restricted.ID) {
		t.Errorf("anonymous GetOpen = %v, must not contain the private liquidation", dtoIDs(anon))
	}
	if !contains(dtoIDs(anon), public.ID) {
		t.Errorf("anonymous GetOpen = %v, want the public liquidation", dtoIDs(anon))
	}

	owner, err := uc.GetOpen(ownerCtx())
	if err != nil {
		t.Fatalf("GetOpen: %v", err)
	}
	if !contains(dtoIDs(owner), restricted.ID) {
		t.Errorf("owner GetOpen = %v, want their own private liquidation", dtoIDs(owner))
	}

	other, err := uc.GetOpen(otherCtx())
	if err != nil {
		t.Fatalf("GetOpen: %v", err)
	}
	if contains(dtoIDs(other), restricted.ID) {
		t.Error("another authenticated user's GetOpen contains the private liquidation")
	}
}

// TestLiquidationUseCaseGetByIDHonoursVisibility is the acceptance gate for the
// single read.
func TestLiquidationUseCaseGetByIDHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	uc := liquidationUC()

	t.Run("anonymous reads a public liquidation", func(t *testing.T) {
		got, err := uc.GetByID(ctx, public.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.ID != public.ID {
			t.Errorf("id = %v, want %v", got.ID, public.ID)
		}
	})

	t.Run("anonymous cannot read a private one", func(t *testing.T) {
		if _, err := uc.GetByID(ctx, restricted.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("another authenticated user cannot read it", func(t *testing.T) {
		if _, err := uc.GetByID(otherCtx(), restricted.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("the owner can read it", func(t *testing.T) {
		got, err := uc.GetByID(ownerCtx(), restricted.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.ID != restricted.ID {
			t.Errorf("id = %v, want %v", got.ID, restricted.ID)
		}
		if got.Visibility != "wholesale" {
			t.Errorf("visibility = %q, want %q", got.Visibility, "wholesale")
		}
	})
}

func TestLiquidationUseCaseGetBySupplierHonoursVisibility(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	public := seedLiquidation(t, liqSupplierID, "Public corn", "public")
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	uc := liquidationUC()

	anon, err := uc.GetBySupplier(ctx, liqSupplierID)
	if err != nil {
		t.Fatalf("GetBySupplier: %v", err)
	}
	if contains(dtoIDs(anon), restricted.ID) {
		t.Error("the anonymous supplier listing contains the private liquidation")
	}
	if !contains(dtoIDs(anon), public.ID) {
		t.Error("the anonymous supplier listing is missing the public liquidation")
	}

	owner, err := uc.GetBySupplier(ownerCtx(), liqSupplierID)
	if err != nil {
		t.Fatalf("GetBySupplier: %v", err)
	}
	if !contains(dtoIDs(owner), restricted.ID) {
		t.Error("the owner supplier listing is missing their own private liquidation")
	}
}

// TestLiquidationMutationsStillReportForbidden pins that the split did not cost
// the mutation paths their error: a non-owner updating somebody else's wholesale
// liquidation is still forbidden, not not-found.
func TestLiquidationMutationsStillReportForbidden(t *testing.T) {
	cleanupTables(t)

	seedLiquidationUser(t, liqSupplierID)
	seedLiquidationUser(t, liqOtherSupID)
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	uc := liquidationUC()

	quantity := 20.0
	if err := uc.UpdateLiquidation(otherCtx(), restricted.ID, dto.UpdateLiquidationRequest{Quantity: &quantity}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("UpdateLiquidation by a non-owner = %v, want %v", err, domain.ErrForbidden)
	}

	if err := uc.DeleteLiquidation(otherCtx(), restricted.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("DeleteLiquidation by a non-owner = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestLiquidationOwnerCanStillMutate(t *testing.T) {
	cleanupTables(t)

	seedLiquidationUser(t, liqSupplierID)
	restricted := seedLiquidation(t, liqSupplierID, "Wholesale corn", "wholesale")

	uc := liquidationUC()

	quantity := 20.0
	if err := uc.UpdateLiquidation(ownerCtx(), restricted.ID, dto.UpdateLiquidationRequest{Quantity: &quantity}); err != nil {
		t.Fatalf("UpdateLiquidation by the owner: %v", err)
	}

	got, err := uc.GetByID(ownerCtx(), restricted.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Quantity != 20 {
		t.Errorf("quantity = %v, want 20", got.Quantity)
	}
}
