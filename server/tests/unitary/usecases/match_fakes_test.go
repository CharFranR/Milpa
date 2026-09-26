package usecases_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
)

var (
	matchTestRequestID    = uuid.MustParse("a1111111-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	matchTestOfferID      = uuid.MustParse("b2222222-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	matchTestOtherOfferID = uuid.MustParse("c3333333-cccc-4ccc-8ccc-cccccccccccc")
	matchTestThirdOfferID = uuid.MustParse("d4444444-dddd-4ddd-8ddd-dddddddddddd")
	matchTestSupplierID   = uuid.MustParse("e5555555-eeee-4eee-8eee-eeeeeeeeeeee")
	matchTestOtherSupply  = uuid.MustParse("f6666666-ffff-4fff-8fff-ffffffffffff")
	matchTestMatchID      = uuid.MustParse("17777777-7777-4777-8777-777777777777")
	matchTestProduct      = "Maize"
)

func matchTestRequest() *domain.SupplyRequest {
	request := domain.NewSupplyRequest(
		testUserID, matchTestProduct, 100, domain.Kg, 1, 100, domain.Kg,
		domain.Address{}, fixedTime, fixedTime.Add(72*time.Hour), "", true,
	)
	request.ID = matchTestRequestID
	return request
}

func matchTestSingleProviderRequest() *domain.SupplyRequest {
	request := matchTestRequest()
	request.MultipleProviders = false
	return request
}

func matchTestOffer() *domain.SupplyOffer {
	offer := domain.NewSupplyOffer(matchTestSupplierID, matchTestRequestID, 30, domain.Kg, fixedTime, true)
	offer.ID = matchTestOfferID
	return offer
}

func matchTestInventory(quantity float32) domain.SupplierInventory {
	return *domain.NewSupplierInventory(matchTestSupplierID, matchTestProduct, quantity, domain.Kg)
}

type fakeMatchSupplyRequestRepo struct {
	getByID func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error)
	update  func(ctx context.Context, request *domain.SupplyRequest) error

	store   map[uuid.UUID]domain.SupplyRequest
	updated []domain.SupplyRequest
}

func newFakeMatchSupplyRequestRepo() *fakeMatchSupplyRequestRepo {
	f := &fakeMatchSupplyRequestRepo{
		store: map[uuid.UUID]domain.SupplyRequest{matchTestRequestID: *matchTestRequest()},
	}
	f.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
		request, ok := f.store[id]
		if !ok {
			return domain.SupplyRequest{}, domain.ErrNotFound
		}
		return request, nil
	}
	f.update = func(ctx context.Context, request *domain.SupplyRequest) error {
		f.store[request.ID] = *request
		f.updated = append(f.updated, *request)
		return nil
	}
	return f
}

func (f *fakeMatchSupplyRequestRepo) Create(ctx context.Context, request *domain.SupplyRequest) error {
	f.store[request.ID] = *request
	return nil
}

func (f *fakeMatchSupplyRequestRepo) List(ctx context.Context, buyerID uuid.UUID) ([]domain.SupplyRequest, error) {
	return nil, nil
}

func (f *fakeMatchSupplyRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
	return f.getByID(ctx, id)
}

func (f *fakeMatchSupplyRequestRepo) Update(ctx context.Context, request *domain.SupplyRequest) error {
	return f.update(ctx, request)
}

func (f *fakeMatchSupplyRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.store, id)
	return nil
}

type fakeMatchSupplyOfferRepo struct {
	getByID       func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error)
	listByRequest func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error)
	update        func(ctx context.Context, offer *domain.SupplyOffer) error

	store   map[uuid.UUID]domain.SupplyOffer
	updated []domain.SupplyOffer
}

func newFakeMatchSupplyOfferRepo() *fakeMatchSupplyOfferRepo {
	f := &fakeMatchSupplyOfferRepo{
		store: map[uuid.UUID]domain.SupplyOffer{matchTestOfferID: *matchTestOffer()},
	}
	f.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
		offer, ok := f.store[id]
		if !ok {
			return domain.SupplyOffer{}, domain.ErrNotFound
		}
		return offer, nil
	}
	f.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
		var offers []domain.SupplyOffer
		for _, offer := range f.store {
			if offer.SupplyRequest == supplyRequestID {
				offers = append(offers, offer)
			}
		}
		return offers, nil
	}
	f.update = func(ctx context.Context, offer *domain.SupplyOffer) error {
		f.store[offer.ID] = *offer
		f.updated = append(f.updated, *offer)
		return nil
	}
	return f
}

func (f *fakeMatchSupplyOfferRepo) seed(offers ...domain.SupplyOffer) {
	for i := range offers {
		f.store[offers[i].ID] = offers[i]
	}
}

func (f *fakeMatchSupplyOfferRepo) Create(ctx context.Context, offer *domain.SupplyOffer) error {
	f.store[offer.ID] = *offer
	return nil
}

func (f *fakeMatchSupplyOfferRepo) List(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplyOffer, error) {
	return nil, nil
}

func (f *fakeMatchSupplyOfferRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
	return f.listByRequest(ctx, supplyRequestID)
}

func (f *fakeMatchSupplyOfferRepo) FindBySupplierAndRequest(ctx context.Context, supplierID, supplyRequestID uuid.UUID) (domain.SupplyOffer, error) {
	for _, offer := range f.store {
		if offer.SupplierID == supplierID && offer.SupplyRequest == supplyRequestID {
			return offer, nil
		}
	}
	return domain.SupplyOffer{}, domain.ErrNotFound
}

func (f *fakeMatchSupplyOfferRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
	return f.getByID(ctx, id)
}

func (f *fakeMatchSupplyOfferRepo) Update(ctx context.Context, offer *domain.SupplyOffer) error {
	return f.update(ctx, offer)
}

func (f *fakeMatchSupplyOfferRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.store, id)
	return nil
}

type fakeMatchRepository struct {
	create                func(ctx context.Context, match *domain.Match) error
	getByID               func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error)
	listByRequest         func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error)
	listActiveByRequest   func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error)
	listActiveBySupplier  func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error)
	existsActiveByRequest func(ctx context.Context, supplyRequestID uuid.UUID) (bool, error)
	existsActiveByOffer   func(ctx context.Context, supplyOfferID uuid.UUID) (bool, error)
	update                func(ctx context.Context, match *domain.Match) error
	delete                func(ctx context.Context, id uuid.UUID) error

	created []*domain.Match
	updated []*domain.Match
	deleted []uuid.UUID
}

func newFakeMatchRepository() *fakeMatchRepository {
	f := &fakeMatchRepository{}
	f.create = func(ctx context.Context, match *domain.Match) error {
		f.created = append(f.created, match)
		return nil
	}
	f.getByID = func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
		return nil, domain.ErrNotFound
	}
	f.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	f.listActiveByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	f.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	f.existsActiveByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
		return false, nil
	}
	f.existsActiveByOffer = func(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
		return false, nil
	}
	f.update = func(ctx context.Context, match *domain.Match) error {
		f.updated = append(f.updated, match)
		return nil
	}
	f.delete = func(ctx context.Context, id uuid.UUID) error {
		f.deleted = append(f.deleted, id)
		return nil
	}
	return f
}

func (f *fakeMatchRepository) Create(ctx context.Context, match *domain.Match) error {
	return f.create(ctx, match)
}

func (f *fakeMatchRepository) ListByOffer(ctx context.Context, supplyOfferID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *fakeMatchRepository) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return f.listByRequest(ctx, supplyRequestID)
}

func (f *fakeMatchRepository) ListActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return f.listActiveByRequest(ctx, supplyRequestID)
}

func (f *fakeMatchRepository) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
	return f.listActiveBySupplier(ctx, supplierID)
}

func (f *fakeMatchRepository) ExistsActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
	return f.existsActiveByRequest(ctx, supplyRequestID)
}

func (f *fakeMatchRepository) ExistsActiveByOffer(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
	return f.existsActiveByOffer(ctx, supplyOfferID)
}

func (f *fakeMatchRepository) GetByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	return f.getByID(ctx, matchID)
}

func (f *fakeMatchRepository) Update(ctx context.Context, match *domain.Match) error {
	return f.update(ctx, match)
}

func (f *fakeMatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return f.delete(ctx, id)
}

type fakeMatchTransactionRepo struct {
	create func(ctx context.Context, transaction *domain.Transaction) error
	update func(ctx context.Context, transaction *domain.Transaction) error
	delete func(ctx context.Context, id uuid.UUID) error

	created []*domain.Transaction
	deleted []uuid.UUID
}

func newFakeMatchTransactionRepo() *fakeMatchTransactionRepo {
	f := &fakeMatchTransactionRepo{}
	f.create = func(ctx context.Context, transaction *domain.Transaction) error {
		f.created = append(f.created, transaction)
		return nil
	}
	f.update = func(ctx context.Context, transaction *domain.Transaction) error {
		return nil
	}
	f.delete = func(ctx context.Context, id uuid.UUID) error {
		f.deleted = append(f.deleted, id)
		return nil
	}
	return f
}

func (f *fakeMatchTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	return f.create(ctx, transaction)
}

func (f *fakeMatchTransactionRepo) List(ctx context.Context, matchID uuid.UUID) ([]domain.Transaction, error) {
	return nil, nil
}

func (f *fakeMatchTransactionRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Transaction, error) {
	return nil, nil
}

func (f *fakeMatchTransactionRepo) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Transaction, error) {
	return nil, nil
}

func (f *fakeMatchTransactionRepo) GetByMatch(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error) {
	return domain.Transaction{}, domain.ErrNotFound
}

func (f *fakeMatchTransactionRepo) GetByID(ctx context.Context, transactionID uuid.UUID) (domain.Transaction, error) {
	return domain.Transaction{}, domain.ErrNotFound
}

func (f *fakeMatchTransactionRepo) Update(ctx context.Context, transaction *domain.Transaction) error {
	return f.update(ctx, transaction)
}

func (f *fakeMatchTransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return f.delete(ctx, id)
}

type fakeMatchInventoryRepo struct {
	findBySupplierAndProduct func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error)

	findCalls []string
}

func newFakeMatchInventoryRepo() *fakeMatchInventoryRepo {
	f := &fakeMatchInventoryRepo{}
	f.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
		return matchTestInventory(100), nil
	}
	return f
}

func (f *fakeMatchInventoryRepo) Create(ctx context.Context, inventory *domain.SupplierInventory) error {
	return nil
}

func (f *fakeMatchInventoryRepo) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplierInventory, error) {
	return nil, nil
}

func (f *fakeMatchInventoryRepo) FindBySupplierAndProduct(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
	f.findCalls = append(f.findCalls, supplierID.String()+"/"+productName)
	return f.findBySupplierAndProduct(ctx, supplierID, productName)
}

func (f *fakeMatchInventoryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.SupplierInventory, error) {
	return domain.SupplierInventory{}, domain.ErrNotFound
}

func (f *fakeMatchInventoryRepo) Update(ctx context.Context, inventory *domain.SupplierInventory) error {
	return nil
}

func (f *fakeMatchInventoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type stubMatchRecommendationUC struct {
	ranked         []*dto.PrioritizedOfferDTO
	err            error
	gotRequestID   uuid.UUID
	gotContext     context.Context
	availableQty   float32
	availableErr   error
	gotSupplierID  uuid.UUID
	gotProductName string
}

func (s *stubMatchRecommendationUC) RankOffers(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error) {
	s.gotRequestID = supplyRequestID
	s.gotContext = ctx
	return s.ranked, s.err
}

func (s *stubMatchRecommendationUC) AvailableQuantity(ctx context.Context, supplierID uuid.UUID, productName string) (float32, error) {
	s.gotSupplierID = supplierID
	s.gotProductName = productName
	return s.availableQty, s.availableErr
}
