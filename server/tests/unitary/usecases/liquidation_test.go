package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

func TestLiquidationUseCaseGetOpenHidesRestrictedFromNonMayoristas(t *testing.T) {
	t.Parallel()

	public := liquidationFixture("Public corn", "public")
	restricted := liquidationFixture("Restricted corn", "private")

	tests := []struct {
		name           string
		ctx            context.Context
		findOpenErr    error
		wantRestricted bool
		wantErr        error
	}{
		{name: "anonymous sees only the public one", ctx: context.Background()},
		{name: "minorista does not see the restricted one", ctx: principalCtx()},
		{name: "an agricultor browsing the market does not see it", ctx: farmerCtx()},
		{name: "mayorista detallista sees it", ctx: mayoristaCtx(), wantRestricted: true},
		{name: "the owning supplier sees its own", ctx: farmerCtxFor(testOtherID), wantRestricted: true},
		{name: "a repository failure is propagated", ctx: principalCtx(), findOpenErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeLiquidationRepo(public, restricted)
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
			if !seen[public.ID] {
				t.Errorf("GetOpen() = %v, want the public liquidation", seen)
			}
			if seen[restricted.ID] != tt.wantRestricted {
				t.Errorf("restricted liquidation visible = %v, want %v", seen[restricted.ID], tt.wantRestricted)
			}
		})
	}
}

func TestLiquidationMutationsRequireANonFarmerToBeRefused(t *testing.T) {
	t.Parallel()

	ownedByOther := liquidationFixture("Restricted corn", "private")

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
