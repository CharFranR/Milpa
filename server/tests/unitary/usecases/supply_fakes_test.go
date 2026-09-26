package usecases_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
)

type supplyFakeRequestRepo struct {
	requests  map[uuid.UUID]domain.SupplyRequest
	created   []*domain.SupplyRequest
	updated   []*domain.SupplyRequest
	listBuyer uuid.UUID
	listErr   error
	getErr    error
	createErr error
	updateErr error
}

func newSupplyFakeRequestRepo() *supplyFakeRequestRepo {
	return &supplyFakeRequestRepo{requests: map[uuid.UUID]domain.SupplyRequest{}}
}

func (f *supplyFakeRequestRepo) Create(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, supplyRequest)
	f.requests[supplyRequest.ID] = *supplyRequest
	return nil
}

func (f *supplyFakeRequestRepo) List(ctx context.Context, buyerID uuid.UUID) ([]domain.SupplyRequest, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	f.listBuyer = buyerID
	var result []domain.SupplyRequest
	for _, supplyRequest := range f.requests {
		if supplyRequest.BuyerID == buyerID {
			result = append(result, supplyRequest)
		}
	}
	return result, nil
}

func (f *supplyFakeRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
	if f.getErr != nil {
		return domain.SupplyRequest{}, f.getErr
	}
	if supplyRequest, ok := f.requests[id]; ok {
		return supplyRequest, nil
	}
	return domain.SupplyRequest{}, domain.ErrNotFound
}

func (f *supplyFakeRequestRepo) Update(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, supplyRequest)
	f.requests[supplyRequest.ID] = *supplyRequest
	return nil
}

func (f *supplyFakeRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.requests, id)
	return nil
}

type supplyFakeOfferRepo struct {
	offers    map[uuid.UUID]domain.SupplyOffer
	created   []*domain.SupplyOffer
	updated   []*domain.SupplyOffer
	findErr   error
	listErr   error
	getErr    error
	createErr error
	updateErr error
}

func newSupplyFakeOfferRepo() *supplyFakeOfferRepo {
	return &supplyFakeOfferRepo{offers: map[uuid.UUID]domain.SupplyOffer{}}
}

func (f *supplyFakeOfferRepo) Create(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, supplyOffer)
	f.offers[supplyOffer.ID] = *supplyOffer
	return nil
}

func (f *supplyFakeOfferRepo) List(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplyOffer, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var result []domain.SupplyOffer
	for _, supplyOffer := range f.offers {
		if supplyOffer.SupplierID == supplierID {
			result = append(result, supplyOffer)
		}
	}
	return result, nil
}

func (f *supplyFakeOfferRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var result []domain.SupplyOffer
	for _, supplyOffer := range f.offers {
		if supplyOffer.SupplyRequest == supplyRequestID {
			result = append(result, supplyOffer)
		}
	}
	return result, nil
}

func (f *supplyFakeOfferRepo) FindBySupplierAndRequest(ctx context.Context, supplierID, supplyRequestID uuid.UUID) (domain.SupplyOffer, error) {
	if f.findErr != nil {
		return domain.SupplyOffer{}, f.findErr
	}
	for _, supplyOffer := range f.offers {
		if supplyOffer.SupplierID == supplierID && supplyOffer.SupplyRequest == supplyRequestID {
			return supplyOffer, nil
		}
	}
	return domain.SupplyOffer{}, domain.ErrNotFound
}

func (f *supplyFakeOfferRepo) GetByID(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
	if f.getErr != nil {
		return domain.SupplyOffer{}, f.getErr
	}
	if supplyOffer, ok := f.offers[supplyOfferID]; ok {
		return supplyOffer, nil
	}
	return domain.SupplyOffer{}, domain.ErrNotFound
}

func (f *supplyFakeOfferRepo) Update(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, supplyOffer)
	f.offers[supplyOffer.ID] = *supplyOffer
	return nil
}

func (f *supplyFakeOfferRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.offers, id)
	return nil
}

type supplyFakeMatchRepo struct {
	existsActive bool
	existsErr    error
}

func newSupplyFakeMatchRepo() *supplyFakeMatchRepo {
	return &supplyFakeMatchRepo{}
}

func (f *supplyFakeMatchRepo) Create(ctx context.Context, match *domain.Match) error {
	return nil
}

func (f *supplyFakeMatchRepo) ListByOffer(ctx context.Context, supplyOfferID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *supplyFakeMatchRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *supplyFakeMatchRepo) ListActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *supplyFakeMatchRepo) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *supplyFakeMatchRepo) ExistsActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
	return f.existsActive, f.existsErr
}

func (f *supplyFakeMatchRepo) ExistsActiveByOffer(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *supplyFakeMatchRepo) GetByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	return nil, domain.ErrNotFound
}

func (f *supplyFakeMatchRepo) Update(ctx context.Context, match *domain.Match) error {
	return nil
}

func (f *supplyFakeMatchRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func supplyTestRequest(buyerID uuid.UUID) domain.SupplyRequest {
	supplyRequest := domain.NewSupplyRequest(
		buyerID, "Rice", 100, domain.Kg, 10, 10, domain.Kg,
		domain.Address{Department: "Masaya", Municipality: "Masaya", AddressLine: "Km 5 Carretera Sur"},
		fixedTime.Add(24*time.Hour), fixedTime.Add(72*time.Hour),
		"Fresh harvest from the cooperative", true,
	)
	supplyRequest.MinAmountPerProvider = 10
	return *supplyRequest
}

func supplyTestOffer(supplierID, supplyRequestID uuid.UUID) domain.SupplyOffer {
	return *domain.NewSupplyOffer(
		supplierID, supplyRequestID, 20, domain.Kg,
		fixedTime.Add(48*time.Hour), true,
	)
}
