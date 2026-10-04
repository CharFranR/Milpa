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

func corporateCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: testUserID, Role: domain.RoleCompradorMayoristaCorporativo})
}

func TestLiquidationUseCaseGetOpenHonoursVisibility(t *testing.T) {
	t.Parallel()

	public := liquidationFixture("Public corn", "public")
	wholesale := liquidationFixture("Wholesale corn", "wholesale")
	retailOnly := liquidationFixture("Retail-only corn", "wholesale_retail")
	corporateOnly := liquidationFixture("Corporate-only corn", "wholesale_corporate")

	tests := []struct {
		name        string
		ctx         context.Context
		findOpenErr error
		want        map[uuid.UUID]bool
		wantErr     error
	}{
		{
			name: "anonymous sees only the public one",
			ctx:  context.Background(),
			want: map[uuid.UUID]bool{public.ID: true},
		},
		{
			name: "minorista sees only the public one",
			ctx:  principalCtx(),
			want: map[uuid.UUID]bool{public.ID: true},
		},
		{
			name: "an agricultor browsing the market sees only the public one",
			ctx:  farmerCtx(),
			want: map[uuid.UUID]bool{public.ID: true},
		},
		{
			name: "mayorista detallista sees wholesale and wholesale_retail but not wholesale_corporate",
			ctx:  mayoristaCtx(),
			want: map[uuid.UUID]bool{public.ID: true, wholesale.ID: true, retailOnly.ID: true},
		},
		{
			name: "mayorista corporativo sees wholesale and wholesale_corporate but not wholesale_retail",
			ctx:  corporateCtx(),
			want: map[uuid.UUID]bool{public.ID: true, wholesale.ID: true, corporateOnly.ID: true},
		},
		{
			name: "the owning supplier sees its own regardless of level",
			ctx:  farmerCtxFor(testOtherID),
			want: map[uuid.UUID]bool{public.ID: true, wholesale.ID: true, retailOnly.ID: true, corporateOnly.ID: true},
		},
		{name: "a repository failure is propagated", ctx: principalCtx(), findOpenErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeLiquidationRepo(public, wholesale, retailOnly, corporateOnly)
			repo.findOpenErr = tt.findOpenErr
			uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

			got, err := uc.GetOpen(tt.ctx)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetOpen() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetOpen() error: %v", err)
			}

			seen := map[uuid.UUID]bool{}
			for _, liquidation := range got {
				seen[liquidation.ID] = true
			}
			for _, liquidation := range []domain.Liquidation{public, wholesale, retailOnly, corporateOnly} {
				if seen[liquidation.ID] != tt.want[liquidation.ID] {
					t.Errorf("liquidation %s (%s) visible = %v, want %v", liquidation.ID, liquidation.Visibility, seen[liquidation.ID], tt.want[liquidation.ID])
				}
			}
		})
	}
}

func TestLiquidationMutationsRequireANonFarmerToBeRefused(t *testing.T) {
	t.Parallel()

	ownedByOther := liquidationFixture("Restricted corn", "wholesale")

	tests := []struct {
		name    string
		ctx     context.Context
		invoke  func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error
		wantErr error
	}{
		{
			name: "create as a minorista",
			ctx:  principalCtx(),
			invoke: func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error {
				_, err := uc.CreateLiquidation(ctx, dto.CreateLiquidationRequest{
					ProductName: "Corn", Quantity: 10, UnitOfMeasure: "kg", TotalPrice: 100, UnitPrice: 10,
				})
				return err
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name: "create as an admin",
			ctx:  reportAdminCtx(),
			invoke: func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error {
				_, err := uc.CreateLiquidation(ctx, dto.CreateLiquidationRequest{
					ProductName: "Corn", Quantity: 10, UnitOfMeasure: "kg", TotalPrice: 100, UnitPrice: 10,
				})
				return err
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name: "update as a minorista",
			ctx:  principalCtx(),
			invoke: func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error {
				quantity := 20.0
				return uc.UpdateLiquidation(ctx, ownedByOther.ID, dto.UpdateLiquidationRequest{Quantity: &quantity})
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name: "delete as a minorista",
			ctx:  principalCtx(),
			invoke: func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error {
				return uc.DeleteLiquidation(ctx, ownedByOther.ID)
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name: "the owning agricultor updates",
			ctx:  farmerCtxFor(testOtherID),
			invoke: func(ctx context.Context, uc *usecases.LiquidationUseCaseImpl) error {
				quantity := 20.0
				return uc.UpdateLiquidation(ctx, ownedByOther.ID, dto.UpdateLiquidationRequest{Quantity: &quantity})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeLiquidationRepo(ownedByOther)
			uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

			err := tt.invoke(tt.ctx, uc)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if len(repo.updated) != 0 || len(repo.deleted) != 0 {
					t.Errorf("a refused caller wrote %d and deleted %d, want nothing", len(repo.updated), len(repo.deleted))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
