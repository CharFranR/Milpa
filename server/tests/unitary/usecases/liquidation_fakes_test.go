package usecases_test

import (
	"context"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

// fakeLiquidationRepo hands the visibility-filtered reads every row it holds,
// without applying the predicate itself. The repository already refuses the
// wrong viewer in its own tests; here the use case is the layer under test and
// has to refuse a restricted liquidation on its own.
type fakeLiquidationRepo struct {
	liquidations []domain.Liquidation
	findOpenErr  error
	findByIDErr  error
	saveErr      error
	updated      []*domain.Liquidation
	deleted      []uuid.UUID
}

func newFakeLiquidationRepo(liquidations ...domain.Liquidation) *fakeLiquidationRepo {
	return &fakeLiquidationRepo{liquidations: liquidations}
}

func (f *fakeLiquidationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Liquidation, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	for i := range f.liquidations {
		if f.liquidations[i].ID == id {
			liq := f.liquidations[i]
			return &liq, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (f *fakeLiquidationRepo) FindVisibleByID(ctx context.Context, id uuid.UUID, viewer port.LiquidationViewer) (*domain.Liquidation, error) {
	return f.FindByID(ctx, id)
}

func (f *fakeLiquidationRepo) FindBySupplier(ctx context.Context, supplierID uuid.UUID, viewer port.LiquidationViewer) ([]domain.Liquidation, error) {
	var result []domain.Liquidation
	for i := range f.liquidations {
		if f.liquidations[i].SupplierID == supplierID {
			result = append(result, f.liquidations[i])
		}
	}
	return result, nil
}

func (f *fakeLiquidationRepo) FindOpen(ctx context.Context, viewer port.LiquidationViewer) ([]domain.Liquidation, error) {
	if f.findOpenErr != nil {
		return nil, f.findOpenErr
	}
	var result []domain.Liquidation
	for i := range f.liquidations {
		if f.liquidations[i].Status == domain.LiquidationOpen {
			result = append(result, f.liquidations[i])
		}
	}
	return result, nil
}

func (f *fakeLiquidationRepo) Save(ctx context.Context, liquidation *domain.Liquidation) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.liquidations = append(f.liquidations, *liquidation)
	return nil
}

func (f *fakeLiquidationRepo) Update(ctx context.Context, liquidation *domain.Liquidation) error {
	f.updated = append(f.updated, liquidation)
	return nil
}

func (f *fakeLiquidationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)
	return nil
}

var _ port.LiquidationRepository = (*fakeLiquidationRepo)(nil)

// liquidationFixture builds an open liquidation owned by a supplier other than
// testUserID, so every case in these tests is decided by the viewer's role.
func liquidationFixture(name, visibility string) domain.Liquidation {
	liquidation, err := domain.NewLiquidation(testOtherID, name, 10, "kg", 100, 10, fixedTime)
	if err != nil {
		panic(err)
	}
	if visibility != "public" {
		if err := liquidation.UpdateVisibility(visibility, fixedTime); err != nil {
			panic(err)
		}
	}
	return *liquidation
}
