package usecases_test

import (
	"context"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type fakeLiquidationRepo struct {
	liquidations    []domain.Liquidation
	interests       []domain.LiquidationInterest
	findOpenErr     error
	findByIDErr     error
	findVisibleErr  error
	saveErr         error
	saveInterestErr error
	updated         []*domain.Liquidation
	deleted         []uuid.UUID
	lastViewer      port.LiquidationViewer
	viewerSeen      bool
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
	f.lastViewer = viewer
	f.viewerSeen = true
	if f.findVisibleErr != nil {
		return nil, f.findVisibleErr
	}
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

func (f *fakeLiquidationRepo) SaveInterest(ctx context.Context, interest *domain.LiquidationInterest) error {
	if f.saveInterestErr != nil {
		return f.saveInterestErr
	}
	for i := range f.interests {
		if f.interests[i].LiquidationID == interest.LiquidationID && f.interests[i].BuyerID == interest.BuyerID {
			return domain.ErrInterestAlreadyExists
		}
	}
	f.interests = append(f.interests, *interest)
	return nil
}

func (f *fakeLiquidationRepo) FindInterests(ctx context.Context, liquidationID uuid.UUID) ([]domain.LiquidationInterest, error) {
	var result []domain.LiquidationInterest
	for i := range f.interests {
		if f.interests[i].LiquidationID == liquidationID {
			result = append(result, f.interests[i])
		}
	}
	return result, nil
}

func (f *fakeLiquidationRepo) InterestExists(ctx context.Context, liquidationID, buyerID uuid.UUID) (bool, error) {
	for i := range f.interests {
		if f.interests[i].LiquidationID == liquidationID && f.interests[i].BuyerID == buyerID {
			return true, nil
		}
	}
	return false, nil
}

var _ port.LiquidationRepository = (*fakeLiquidationRepo)(nil)

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
