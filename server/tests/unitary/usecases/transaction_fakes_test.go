package usecases_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

var (
	txTestMatchID      = uuid.MustParse("d1111111-1111-4111-8111-111111111111")
	txTestMatchTwoID   = uuid.MustParse("d2222222-2222-4222-8222-222222222222")
	txTestTransactID   = uuid.MustParse("d3333333-3333-4333-8333-333333333333")
	txTestTransacTwoID = uuid.MustParse("d4444444-4444-4444-8444-444444444444")
	txTestRequestID    = uuid.MustParse("d5555555-5555-4555-8555-555555555555")
	txTestOfferID      = uuid.MustParse("d6666666-6666-4666-8666-666666666666")
	txTestOfferTwoID   = uuid.MustParse("d7777777-7777-4777-8777-777777777777")
	txTestBuyerID      = uuid.MustParse("d8888888-8888-4888-8888-888888888888")
	txTestSupplierID   = uuid.MustParse("d9999999-9999-4999-8999-999999999999")
	txTestSupplier2ID  = uuid.MustParse("deaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
)

func txMustTransaction() *domain.Transaction {
	transaction := domain.NewTransaction(txTestMatchID)
	transaction.ID = txTestTransactID
	transaction.CreatedAt = fixedTime
	transaction.UpdatedAt = fixedTime
	return transaction
}

func txMustMatch() *domain.Match {
	return &domain.Match{
		ID:            txTestMatchID,
		SupplyOffer:   txTestOfferID,
		SupplyRequest: txTestRequestID,
		Status:        domain.MatchActive,
		MatchedAmount: 40,
		CreatedAt:     fixedTime,
		UpdatedAt:     fixedTime,
	}
}

func txMustRequest() *domain.SupplyRequest {
	return &domain.SupplyRequest{
		ID:           txTestRequestID,
		BuyerID:      txTestBuyerID,
		ProductName:  "Organic Corn",
		TotalAmount:  100,
		ActualAmount: 60,
		Status:       domain.SupplyRequestOpen,
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}
}

func txMustOffer() *domain.SupplyOffer {
	return &domain.SupplyOffer{
		ID:            txTestOfferID,
		SupplierID:    txTestSupplierID,
		SupplyRequest: txTestRequestID,
		TotalAmount:   40,
		Status:        domain.OfferMatched,
		CreatedAt:     fixedTime,
		UpdatedAt:     fixedTime,
	}
}

type fakeTxTransactionRepo struct {
	getByID       func(ctx context.Context, id uuid.UUID) (domain.Transaction, error)
	getByMatch    func(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error)
	listByRequest func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Transaction, error)
	update        func(ctx context.Context, transaction *domain.Transaction) error
	stored        domain.Transaction
	updated       []*domain.Transaction
}

func newFakeTxTransactionRepo() *fakeTxTransactionRepo {
	f := &fakeTxTransactionRepo{stored: *txMustTransaction()}
	f.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
		return f.stored, nil
	}
	f.getByMatch = func(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error) {
		return f.stored, nil
	}
	f.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Transaction, error) {
		return nil, nil
	}
	f.update = func(ctx context.Context, transaction *domain.Transaction) error {
		f.stored = *transaction
		f.updated = append(f.updated, transaction)
		return nil
	}
	return f
}

func (f *fakeTxTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	return nil
}

func (f *fakeTxTransactionRepo) List(ctx context.Context, matchID uuid.UUID) ([]domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTxTransactionRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Transaction, error) {
	return f.listByRequest(ctx, supplyRequestID)
}

func (f *fakeTxTransactionRepo) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTxTransactionRepo) GetByMatch(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error) {
	return f.getByMatch(ctx, matchID)
}

func (f *fakeTxTransactionRepo) GetByID(ctx context.Context, transactionID uuid.UUID) (domain.Transaction, error) {
	return f.getByID(ctx, transactionID)
}

func (f *fakeTxTransactionRepo) Update(ctx context.Context, transaction *domain.Transaction) error {
	return f.update(ctx, transaction)
}

func (f *fakeTxTransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (f *fakeTxTransactionRepo) LockByIDForUpdate(ctx context.Context, transactionID uuid.UUID) (domain.Transaction, error) {
	return f.getByID(ctx, transactionID)
}

type fakeTxMatchRepo struct {
	getByID       func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error)
	listByRequest func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error)
	update        func(ctx context.Context, match *domain.Match) error
	stored        domain.Match
	updated       []*domain.Match
}

func newFakeTxMatchRepo() *fakeTxMatchRepo {
	f := &fakeTxMatchRepo{stored: *txMustMatch()}
	f.getByID = func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
		match := f.stored
		return &match, nil
	}
	f.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	f.update = func(ctx context.Context, match *domain.Match) error {
		f.stored = *match
		f.updated = append(f.updated, match)
		return nil
	}
	return f
}

func (f *fakeTxMatchRepo) Create(ctx context.Context, match *domain.Match) error {
	return nil
}

func (f *fakeTxMatchRepo) ListByOffer(ctx context.Context, supplyOfferID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *fakeTxMatchRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return f.listByRequest(ctx, supplyRequestID)
}

func (f *fakeTxMatchRepo) ListActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *fakeTxMatchRepo) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
	return nil, nil
}

func (f *fakeTxMatchRepo) ExistsActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeTxMatchRepo) ExistsActiveByOffer(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeTxMatchRepo) GetByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	return f.getByID(ctx, matchID)
}

func (f *fakeTxMatchRepo) Update(ctx context.Context, match *domain.Match) error {
	return f.update(ctx, match)
}

func (f *fakeTxMatchRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (f *fakeTxMatchRepo) LockByIDForUpdate(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	return f.getByID(ctx, matchID)
}

type fakeTxRequestRepo struct {
	getByID func(ctx context.Context, supplyRequestID uuid.UUID) (domain.SupplyRequest, error)
	update  func(ctx context.Context, supplyRequest *domain.SupplyRequest) error
	stored  domain.SupplyRequest
	updated []*domain.SupplyRequest
}

func newFakeTxRequestRepo() *fakeTxRequestRepo {
	f := &fakeTxRequestRepo{stored: *txMustRequest()}
	f.getByID = func(ctx context.Context, supplyRequestID uuid.UUID) (domain.SupplyRequest, error) {
		return f.stored, nil
	}
	f.update = func(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
		f.stored = *supplyRequest
		f.updated = append(f.updated, supplyRequest)
		return nil
	}
	return f
}

func (f *fakeTxRequestRepo) Create(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	return nil
}

func (f *fakeTxRequestRepo) List(ctx context.Context, buyerID uuid.UUID) ([]domain.SupplyRequest, error) {
	return nil, nil
}

func (f *fakeTxRequestRepo) ListOpen(ctx context.Context) ([]domain.SupplyRequest, error) {
	return nil, nil
}

func (f *fakeTxRequestRepo) GetByID(ctx context.Context, supplyRequestID uuid.UUID) (domain.SupplyRequest, error) {
	return f.getByID(ctx, supplyRequestID)
}

func (f *fakeTxRequestRepo) Update(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	return f.update(ctx, supplyRequest)
}

func (f *fakeTxRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// LockForUpdate delegates to the same swappable read as GetByID: the fake has
// no rows to lock, and a test that swaps getByID expects the locked read to
// observe the same value.
func (f *fakeTxRequestRepo) LockForUpdate(ctx context.Context, supplyRequestID uuid.UUID) (domain.SupplyRequest, error) {
	return f.getByID(ctx, supplyRequestID)
}

// Reserve mirrors the SQL predicate (actual_amount >= amount) and then reports
// through the same swappable write as Update.
func (f *fakeTxRequestRepo) Reserve(ctx context.Context, supplyRequestID uuid.UUID, amount float64, at time.Time) error {
	current, err := f.getByID(ctx, supplyRequestID)
	if err != nil {
		return err
	}
	if current.ActualAmount < amount {
		return fmt.Errorf("fakeTxRequestRepo.Reserve: %w", domain.ErrInsufficientAmount)
	}
	current.ActualAmount -= amount
	current.UpdatedAt = at
	return f.update(ctx, &current)
}

func (f *fakeTxRequestRepo) Release(ctx context.Context, supplyRequestID uuid.UUID, amount float64, at time.Time) error {
	current, err := f.getByID(ctx, supplyRequestID)
	if err != nil {
		return err
	}
	current.ActualAmount = min(current.ActualAmount+amount, current.TotalAmount)
	current.UpdatedAt = at
	return f.update(ctx, &current)
}

func (f *fakeTxRequestRepo) UpdateStatus(ctx context.Context, supplyRequestID uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error {
	current, err := f.getByID(ctx, supplyRequestID)
	if err != nil {
		return err
	}
	current.Status = status
	current.UpdatedAt = at
	return f.update(ctx, &current)
}

type fakeTxOfferRepo struct {
	getByID func(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error)
	update  func(ctx context.Context, supplyOffer *domain.SupplyOffer) error
	stored  domain.SupplyOffer
	updated []*domain.SupplyOffer
}

func newFakeTxOfferRepo() *fakeTxOfferRepo {
	f := &fakeTxOfferRepo{stored: *txMustOffer()}
	f.getByID = func(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
		return f.stored, nil
	}
	f.update = func(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
		f.stored = *supplyOffer
		f.updated = append(f.updated, supplyOffer)
		return nil
	}
	return f
}

func (f *fakeTxOfferRepo) Create(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	return nil
}

func (f *fakeTxOfferRepo) List(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplyOffer, error) {
	return nil, nil
}

func (f *fakeTxOfferRepo) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
	return nil, nil
}

func (f *fakeTxOfferRepo) FindBySupplierAndRequest(ctx context.Context, supplierID, supplyRequestID uuid.UUID) (domain.SupplyOffer, error) {
	return domain.SupplyOffer{}, nil
}

func (f *fakeTxOfferRepo) GetByID(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
	return f.getByID(ctx, supplyOfferID)
}

func (f *fakeTxOfferRepo) Update(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	return f.update(ctx, supplyOffer)
}

func (f *fakeTxOfferRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (f *fakeTxOfferRepo) LockByIDForUpdate(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
	return f.getByID(ctx, supplyOfferID)
}

// newTxScope wires the four fakes of the transaction fixture into the scope the
// use case receives. The fakes themselves already satisfy the widened
// tx-scoped ports, so no adapter type is needed.
func (f *txFixture) newTxScope() port.TxScope {
	return port.TxScope{
		Requests:     f.requestRepo,
		Offers:       f.offerRepo,
		Matches:      f.matchRepo,
		Transactions: f.transactionRepo,
	}
}

var (
	_ port.TransactionRepository   = (*fakeTxTransactionRepo)(nil)
	_ port.MatchRepository         = (*fakeTxMatchRepo)(nil)
	_ port.SupplyRequestRepository = (*fakeTxRequestRepo)(nil)
	_ port.SupplyOfferRepository   = (*fakeTxOfferRepo)(nil)
	_ port.RequestReservationStore = (*fakeTxRequestRepo)(nil)
	_ port.TxTransactionRepository = (*fakeTxTransactionRepo)(nil)
	_ port.TxMatchRepository       = (*fakeTxMatchRepo)(nil)
	_ port.TxOfferRepository       = (*fakeTxOfferRepo)(nil)
)
