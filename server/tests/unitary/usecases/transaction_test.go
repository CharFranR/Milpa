package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

type txFixture struct {
	transactionRepo *fakeTxTransactionRepo
	matchRepo       *fakeTxMatchRepo
	requestRepo     *fakeTxRequestRepo
	offerRepo       *fakeTxOfferRepo
	uow             *fakeUnitOfWork
	uc              *usecases.TransactionUseCaseImpl
}

func newTxFixture() *txFixture {
	transactionRepo := newFakeTxTransactionRepo()
	matchRepo := newFakeTxMatchRepo()
	requestRepo := newFakeTxRequestRepo()
	offerRepo := newFakeTxOfferRepo()
	f := &txFixture{
		transactionRepo: transactionRepo,
		matchRepo:       matchRepo,
		requestRepo:     requestRepo,
		offerRepo:       offerRepo,
	}
	f.uow = newFakeUnitOfWork(f.newTxScope())
	f.uc = usecases.NewTransactionUseCase(
		transactionRepo,
		matchRepo,
		requestRepo,
		offerRepo,
		newFakeTimer(),
		f.uow,
	)
	return f
}

func txInProgressTransaction() domain.Transaction {
	transaction := *txMustTransaction()
	transaction.Status = domain.TransactionInProgress
	buyerStart := fixedTime
	supplierStart := fixedTime.Add(time.Minute)
	transaction.BuyerStartConfirmedAt = &buyerStart
	transaction.SupplierStartConfirmedAt = &supplierStart
	return transaction
}

func txCompletedTransaction() domain.Transaction {
	transaction := txInProgressTransaction()
	transaction.Status = domain.TransactionCompleted
	buyerDelivery := fixedTime.Add(2 * time.Minute)
	supplierDelivery := fixedTime.Add(3 * time.Minute)
	transaction.BuyerDeliveryConfirmedAt = &buyerDelivery
	transaction.SupplierDeliveryConfirmedAt = &supplierDelivery
	return transaction
}

func TestTransactionUseCaseGetByMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ctx            context.Context
		matchErr       error
		requestErr     error
		offerErr       error
		transactionErr error
		wantErr        error
	}{
		{name: "buyer can read own transaction", ctx: principalCtxFor(txTestBuyerID)},
		{name: "supplier can read own transaction", ctx: principalCtxFor(txTestSupplierID)},
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "match not found", ctx: principalCtxFor(txTestBuyerID), matchErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "request not found", ctx: principalCtxFor(txTestBuyerID), requestErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "offer not found", ctx: principalCtxFor(txTestBuyerID), offerErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "third party forbidden", ctx: principalCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "transaction not found", ctx: principalCtxFor(txTestBuyerID), transactionErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "authorization checked before transaction load", ctx: principalCtxFor(testOtherID), transactionErr: errFake, wantErr: domain.ErrForbidden},
		{name: "repository error propagated", ctx: principalCtxFor(txTestBuyerID), matchErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			if tt.matchErr != nil {
				fx.matchRepo.getByID = func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
					return nil, tt.matchErr
				}
			}
			if tt.requestErr != nil {
				fx.requestRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
					return domain.SupplyRequest{}, tt.requestErr
				}
			}
			if tt.offerErr != nil {
				fx.offerRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
					return domain.SupplyOffer{}, tt.offerErr
				}
			}
			if tt.transactionErr != nil {
				fx.transactionRepo.getByMatch = func(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error) {
					return domain.Transaction{}, tt.transactionErr
				}
			}

			got, err := fx.uc.GetByMatch(tt.ctx, txTestMatchID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("expected nil transaction, got %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected transaction, got nil")
			}
			if got.MatchID == nil || *got.MatchID != txTestMatchID {
				t.Errorf("match id = %v, want %v", got.MatchID, txTestMatchID)
			}
			if len(got.History) == 0 || got.History[0].Status != domain.TransactionMatched {
				t.Errorf("expected history to start at matched, got %+v", got.History)
			}
		})
	}
}

func TestTransactionUseCaseListByRequest(t *testing.T) {
	t.Parallel()

	listScenario := func() ([]domain.Transaction, []domain.Match) {
		first := *txMustTransaction()
		first.MatchID = txTestMatchID
		second := *txMustTransaction()
		second.ID = txTestTransacTwoID
		second.MatchID = txTestMatchTwoID
		matches := []domain.Match{
			{ID: txTestMatchID, SupplyOffer: txTestOfferID, SupplyRequest: txTestRequestID, MatchedAmount: 40, Status: domain.MatchActive},
			{ID: txTestMatchTwoID, SupplyOffer: txTestOfferTwoID, SupplyRequest: txTestRequestID, MatchedAmount: 60, Status: domain.MatchActive},
		}
		return []domain.Transaction{first, second}, matches
	}

	tests := []struct {
		name         string
		ctx          context.Context
		transactions []domain.Transaction
		requestErr   error
		listErr      error
		matchListErr error
		offerErr     error
		wantCount    int
		wantErr      error
	}{
		{
			name:         "buyer sees all transactions of the request",
			ctx:          principalCtxFor(txTestBuyerID),
			transactions: func() []domain.Transaction { txs, _ := listScenario(); return txs }(),
			wantCount:    2,
		},
		{
			name:         "supplier sees only own transactions",
			ctx:          principalCtxFor(txTestSupplierID),
			transactions: func() []domain.Transaction { txs, _ := listScenario(); return txs }(),
			wantCount:    1,
		},
		{
			name:         "second supplier sees only own transactions",
			ctx:          principalCtxFor(txTestSupplier2ID),
			transactions: func() []domain.Transaction { txs, _ := listScenario(); return txs }(),
			wantCount:    1,
		},
		{
			name: "third party is forbidden",
			ctx:  principalCtxFor(testOtherID),
			transactions: func() []domain.Transaction {
				txs, _ := listScenario()
				return txs
			}(),
			wantErr: domain.ErrForbidden,
		},
		{
			name: "supplier without transactions on the request is forbidden",
			ctx:  principalCtxFor(txTestSupplierID),
			transactions: func() []domain.Transaction {
				only := *txMustTransaction()
				only.MatchID = txTestMatchTwoID
				return []domain.Transaction{only}
			}(),
			wantErr: domain.ErrForbidden,
		},
		{
			name:         "buyer with no transactions gets empty list",
			ctx:          principalCtxFor(txTestBuyerID),
			transactions: []domain.Transaction{},
			wantCount:    0,
		},
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "request not found", ctx: principalCtxFor(txTestBuyerID), requestErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "transaction list error", ctx: principalCtxFor(txTestBuyerID), listErr: errFake, wantErr: errFake},
		{name: "match list error for non buyer", ctx: principalCtxFor(txTestSupplierID), matchListErr: errFake, wantErr: errFake},
		{name: "offer lookup error", ctx: principalCtxFor(txTestSupplierID), offerErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			transactions, matches := listScenario()
			if tt.transactions != nil {
				transactions = tt.transactions
			}
			fx.transactionRepo.listByRequest = func(ctx context.Context, id uuid.UUID) ([]domain.Transaction, error) {
				if tt.listErr != nil {
					return nil, tt.listErr
				}
				return transactions, nil
			}
			fx.matchRepo.listByRequest = func(ctx context.Context, id uuid.UUID) ([]domain.Match, error) {
				if tt.matchListErr != nil {
					return nil, tt.matchListErr
				}
				return matches, nil
			}
			fx.requestRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
				if tt.requestErr != nil {
					return domain.SupplyRequest{}, tt.requestErr
				}
				return *txMustRequest(), nil
			}
			fx.offerRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
				if tt.offerErr != nil {
					return domain.SupplyOffer{}, tt.offerErr
				}
				offer := *txMustOffer()
				if id == txTestOfferTwoID {
					offer.ID = txTestOfferTwoID
					offer.SupplierID = txTestSupplier2ID
				}
				return offer, nil
			}

			got, err := fx.uc.ListByRequest(tt.ctx, txTestRequestID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("expected nil list, got %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Fatalf("len(got) = %d, want %d (%+v)", len(got), tt.wantCount, got)
			}
		})
	}
}

func TestTransactionUseCaseConfirmStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		getErr     error
		matchErr   error
		requestErr error
		offerErr   error
		updateErr  error
		wantErr    error
		wantStatus domain.TransactionStatus
	}{
		{
			name:       "buyer first confirmation keeps matched",
			ctx:        principalCtxFor(txTestBuyerID),
			wantStatus: domain.TransactionMatched,
		},
		{
			name:       "supplier can confirm start",
			ctx:        principalCtxFor(txTestSupplierID),
			wantStatus: domain.TransactionMatched,
		},
		{
			name:    "unauthenticated",
			ctx:     context.Background(),
			wantErr: auth.ErrUnauthenticated,
		},
		{
			name:    "transaction not found",
			ctx:     principalCtxFor(txTestBuyerID),
			getErr:  domain.ErrNotFound,
			wantErr: domain.ErrNotFound,
		},
		{
			name:     "match not found",
			ctx:      principalCtxFor(txTestBuyerID),
			matchErr: domain.ErrNotFound,
			wantErr:  domain.ErrNotFound,
		},
		{
			name:       "request not found",
			ctx:        principalCtxFor(txTestBuyerID),
			requestErr: domain.ErrNotFound,
			wantErr:    domain.ErrNotFound,
		},
		{
			name:     "offer not found",
			ctx:      principalCtxFor(txTestBuyerID),
			offerErr: domain.ErrNotFound,
			wantErr:  domain.ErrNotFound,
		},
		{
			name:    "third party forbidden",
			ctx:     principalCtxFor(testOtherID),
			wantErr: domain.ErrForbidden,
		},
		{
			name:      "repository update error propagated",
			ctx:       principalCtxFor(txTestBuyerID),
			updateErr: errFake,
			wantErr:   errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			if tt.getErr != nil {
				fx.transactionRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
					return domain.Transaction{}, tt.getErr
				}
			}
			if tt.matchErr != nil {
				fx.matchRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
					return nil, tt.matchErr
				}
			}
			if tt.requestErr != nil {
				fx.requestRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
					return domain.SupplyRequest{}, tt.requestErr
				}
			}
			if tt.offerErr != nil {
				fx.offerRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
					return domain.SupplyOffer{}, tt.offerErr
				}
			}
			if tt.updateErr != nil {
				fx.transactionRepo.update = func(ctx context.Context, transaction *domain.Transaction) error {
					return tt.updateErr
				}
			}

			err := fx.uc.ConfirmStart(tt.ctx, txTestTransactID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if len(fx.transactionRepo.updated) != 0 {
					t.Fatalf("expected no persistence, got %d updates", len(fx.transactionRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fx.transactionRepo.updated) != 1 {
				t.Fatalf("expected 1 update, got %d", len(fx.transactionRepo.updated))
			}
			if fx.transactionRepo.stored.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", fx.transactionRepo.stored.Status, tt.wantStatus)
			}
		})
	}
}

func TestTransactionUseCaseConfirmStartProgression(t *testing.T) {
	t.Parallel()

	fx := newTxFixture()

	if err := fx.uc.ConfirmStart(principalCtxFor(txTestBuyerID), txTestTransactID); err != nil {
		t.Fatalf("buyer confirm: %v", err)
	}
	if fx.transactionRepo.stored.Status != domain.TransactionMatched {
		t.Fatalf("status after first confirm = %v, want matched", fx.transactionRepo.stored.Status)
	}
	if fx.transactionRepo.stored.BuyerStartConfirmedAt == nil {
		t.Fatal("expected buyer start confirmation timestamp")
	}
	if fx.transactionRepo.stored.SupplierStartConfirmedAt != nil {
		t.Fatal("expected no supplier start confirmation yet")
	}

	if err := fx.uc.ConfirmStart(principalCtxFor(txTestSupplierID), txTestTransactID); err != nil {
		t.Fatalf("supplier confirm: %v", err)
	}
	if fx.transactionRepo.stored.Status != domain.TransactionInProgress {
		t.Fatalf("status after second confirm = %v, want in progress", fx.transactionRepo.stored.Status)
	}
	if fx.transactionRepo.stored.SupplierStartConfirmedAt == nil {
		t.Fatal("expected supplier start confirmation timestamp")
	}
	if len(fx.transactionRepo.updated) != 2 {
		t.Fatalf("expected 2 persisted updates, got %d", len(fx.transactionRepo.updated))
	}
}

func TestTransactionUseCaseConfirmStartRejectsDoubleConfirm(t *testing.T) {
	t.Parallel()

	fx := newTxFixture()

	if err := fx.uc.ConfirmStart(principalCtxFor(txTestBuyerID), txTestTransactID); err != nil {
		t.Fatalf("first confirm: %v", err)
	}

	err := fx.uc.ConfirmStart(principalCtxFor(txTestBuyerID), txTestTransactID)

	if !errors.Is(err, domain.ErrAlreadyConfirmed) {
		t.Fatalf("error = %v, want %v", err, domain.ErrAlreadyConfirmed)
	}
	if len(fx.transactionRepo.updated) != 1 {
		t.Fatalf("expected no extra persistence, got %d updates", len(fx.transactionRepo.updated))
	}
}

func TestTransactionUseCaseConfirmDelivery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		prepare    func(fx *txFixture)
		getErr     error
		updateErr  error
		wantErr    error
		wantStatus domain.TransactionStatus
	}{
		{
			name: "first delivery confirmation keeps in progress",
			ctx:  principalCtxFor(txTestBuyerID),
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = txInProgressTransaction()
				fx.transactionRepo.listByRequest = func(ctx context.Context, id uuid.UUID) ([]domain.Transaction, error) {
					return nil, errFake
				}
			},
			wantStatus: domain.TransactionInProgress,
		},
		{
			name: "delivery rejected while matched",
			ctx:  principalCtxFor(txTestBuyerID),
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = *txMustTransaction()
			},
			wantErr: domain.ErrInvalidTransactionTransition,
		},
		{
			name: "double delivery confirmation rejected",
			ctx:  principalCtxFor(txTestBuyerID),
			prepare: func(fx *txFixture) {
				transaction := txInProgressTransaction()
				buyerDelivery := fixedTime.Add(2 * time.Minute)
				transaction.BuyerDeliveryConfirmedAt = &buyerDelivery
				fx.transactionRepo.stored = transaction
			},
			wantErr: domain.ErrAlreadyConfirmed,
		},
		{
			name: "third party forbidden",
			ctx:  principalCtxFor(testOtherID),
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = txInProgressTransaction()
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name:    "unauthenticated",
			ctx:     context.Background(),
			wantErr: auth.ErrUnauthenticated,
		},
		{
			name: "transaction not found",
			ctx:  principalCtxFor(txTestBuyerID),
			prepare: func(fx *txFixture) {
				fx.transactionRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
					return domain.Transaction{}, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "repository update error propagated",
			ctx:  principalCtxFor(txTestBuyerID),
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = txInProgressTransaction()
				fx.transactionRepo.update = func(ctx context.Context, transaction *domain.Transaction) error {
					return errFake
				}
			},
			wantErr: errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			if tt.prepare != nil {
				tt.prepare(fx)
			}
			if tt.getErr != nil {
				fx.transactionRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
					return domain.Transaction{}, tt.getErr
				}
			}

			err := fx.uc.ConfirmDelivery(tt.ctx, txTestTransactID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if len(fx.transactionRepo.updated) != 0 {
					t.Fatalf("expected no persistence, got %d updates", len(fx.transactionRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fx.transactionRepo.stored.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", fx.transactionRepo.stored.Status, tt.wantStatus)
			}
		})
	}
}

func TestTransactionUseCaseConfirmDeliveryCompletion(t *testing.T) {
	t.Parallel()

	fx := newTxFixture()
	transaction := txInProgressTransaction()
	transaction.BuyerDeliveryConfirmedAt = nil
	fx.transactionRepo.stored = transaction

	if err := fx.uc.ConfirmDelivery(principalCtxFor(txTestBuyerID), txTestTransactID); err != nil {
		t.Fatalf("first delivery confirm: %v", err)
	}
	if fx.transactionRepo.stored.Status != domain.TransactionInProgress {
		t.Fatalf("status after first delivery = %v, want in progress", fx.transactionRepo.stored.Status)
	}

	if err := fx.uc.ConfirmDelivery(principalCtxFor(txTestSupplierID), txTestTransactID); err != nil {
		t.Fatalf("second delivery confirm: %v", err)
	}
	if fx.transactionRepo.stored.Status != domain.TransactionCompleted {
		t.Fatalf("status after second delivery = %v, want completed", fx.transactionRepo.stored.Status)
	}
}

func TestTransactionUseCaseAutoCloseRequest(t *testing.T) {
	t.Parallel()

	completedTx := func(matchID uuid.UUID) domain.Transaction {
		transaction := txCompletedTransaction()
		transaction.MatchID = matchID
		return transaction
	}

	tests := []struct {
		name             string
		totalAmount      float64
		actualAmount     float64
		requestStatus    domain.SupplyRequestStatus
		requestList      []domain.Transaction
		matchList        []domain.Match
		listErr          error
		matchListErr     error
		requestUpdateErr error
		wantClosed       bool
		wantErr          error
	}{
		{
			name:         "completed sum reaches total closes request",
			totalAmount:  40,
			actualAmount: 0,
			requestList:  []domain.Transaction{completedTx(txTestMatchID)},
			matchList:    []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			wantClosed:   true,
		},
		{
			name:         "completed sum below total keeps request open",
			totalAmount:  100,
			actualAmount: 0,
			requestList:  []domain.Transaction{completedTx(txTestMatchID)},
			matchList:    []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			wantClosed:   false,
		},
		{
			name:         "actual amount is not the completion measure",
			totalAmount:  100,
			actualAmount: 0,
			requestList:  []domain.Transaction{completedTx(txTestMatchID)},
			matchList:    []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			wantClosed:   false,
		},
		{
			name:          "non open request is not completed again",
			totalAmount:   40,
			actualAmount:  0,
			requestStatus: domain.SupplyRequestCancelled,
			requestList:   []domain.Transaction{completedTx(txTestMatchID)},
			matchList:     []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			wantClosed:    false,
		},
		{
			name:         "completed transaction without matching match contributes zero",
			totalAmount:  40,
			actualAmount: 0,
			requestList:  []domain.Transaction{completedTx(uuid.New())},
			matchList:    []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			wantClosed:   false,
		},
		{
			name:        "transaction list error propagated",
			totalAmount: 40,
			requestList: nil,
			listErr:     errFake,
			wantErr:     errFake,
		},
		{
			name:         "match list error propagated",
			totalAmount:  40,
			requestList:  []domain.Transaction{completedTx(txTestMatchID)},
			matchListErr: errFake,
			wantErr:      errFake,
		},
		{
			name:             "request update error propagated",
			totalAmount:      40,
			requestList:      []domain.Transaction{completedTx(txTestMatchID)},
			matchList:        []domain.Match{{ID: txTestMatchID, MatchedAmount: 40}},
			requestUpdateErr: errFake,
			wantErr:          errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			transaction := txInProgressTransaction()
			buyerDelivery := fixedTime.Add(2 * time.Minute)
			transaction.BuyerDeliveryConfirmedAt = &buyerDelivery
			fx.transactionRepo.stored = transaction

			request := *txMustRequest()
			request.TotalAmount = tt.totalAmount
			request.ActualAmount = tt.actualAmount
			request.Status = tt.requestStatus
			fx.requestRepo.stored = request
			if tt.requestUpdateErr != nil {
				fx.requestRepo.update = func(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
					return tt.requestUpdateErr
				}
			}
			fx.transactionRepo.listByRequest = func(ctx context.Context, id uuid.UUID) ([]domain.Transaction, error) {
				if tt.listErr != nil {
					return nil, tt.listErr
				}
				return tt.requestList, nil
			}
			fx.matchRepo.listByRequest = func(ctx context.Context, id uuid.UUID) ([]domain.Match, error) {
				if tt.matchListErr != nil {
					return nil, tt.matchListErr
				}
				return tt.matchList, nil
			}

			err := fx.uc.ConfirmDelivery(principalCtxFor(txTestSupplierID), txTestTransactID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			closed := len(fx.requestRepo.updated) > 0
			if closed != tt.wantClosed {
				t.Fatalf("request closed = %v, want %v (updates: %d)", closed, tt.wantClosed, len(fx.requestRepo.updated))
			}
			if tt.wantClosed {
				if fx.requestRepo.updated[0].Status != domain.SupplyRequestCompleted {
					t.Errorf("request status = %v, want completed", fx.requestRepo.updated[0].Status)
				}
			}
		})
	}
}

func TestTransactionUseCaseCancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		ctx              context.Context
		reason           string
		prepare          func(fx *txFixture)
		getErr           error
		matchErr         error
		updateErr        error
		matchUpdateErr   error
		requestUpdateErr error
		offerUpdateErr   error
		offerUntouched   bool
		wantErr          error
		wantCascade      bool
		wantCancelledBy  *uuid.UUID
	}{
		{
			name:            "buyer cancels matched transaction with cascade",
			ctx:             principalCtxFor(txTestBuyerID),
			reason:          "supplier missed the delivery window",
			wantCascade:     true,
			wantCancelledBy: &txTestBuyerID,
		},
		{
			name:   "supplier cancels in progress transaction",
			ctx:    principalCtxFor(txTestSupplierID),
			reason: "buyer unreachable",
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = txInProgressTransaction()
			},
			wantCascade:     true,
			wantCancelledBy: &txTestSupplierID,
		},
		{
			name:    "empty reason rejected",
			ctx:     principalCtxFor(txTestBuyerID),
			reason:  "",
			wantErr: domain.ErrReasonRequired,
		},
		{
			name:   "completed transaction is terminal",
			ctx:    principalCtxFor(txTestBuyerID),
			reason: "too late",
			prepare: func(fx *txFixture) {
				fx.transactionRepo.stored = txCompletedTransaction()
			},
			wantErr: domain.ErrTerminalState,
		},
		{
			name:   "cancelled transaction is terminal",
			ctx:    principalCtxFor(txTestBuyerID),
			reason: "again",
			prepare: func(fx *txFixture) {
				transaction := *txMustTransaction()
				transaction.Status = domain.TransactionCancelled
				fx.transactionRepo.stored = transaction
			},
			wantErr: domain.ErrTerminalState,
		},
		{
			name:    "third party forbidden",
			ctx:     principalCtxFor(testOtherID),
			reason:  "not mine",
			wantErr: domain.ErrForbidden,
		},
		{
			name:    "unauthenticated",
			ctx:     context.Background(),
			reason:  "no token",
			wantErr: auth.ErrUnauthenticated,
		},
		{
			name:    "transaction not found",
			ctx:     principalCtxFor(txTestBuyerID),
			reason:  "missing",
			getErr:  domain.ErrNotFound,
			wantErr: domain.ErrNotFound,
		},
		{
			name:     "match not found",
			ctx:      principalCtxFor(txTestBuyerID),
			reason:   "missing",
			matchErr: domain.ErrNotFound,
			wantErr:  domain.ErrNotFound,
		},
		{
			name:   "already cancelled match rejected before persistence",
			ctx:    principalCtxFor(txTestBuyerID),
			reason: "conflict",
			prepare: func(fx *txFixture) {
				match := txMustMatch()
				match.Status = domain.MatchCancelled
				fx.matchRepo.stored = *match
			},
			wantErr: domain.ErrInvalidMatchStatus,
		},
		{
			name:   "zero matched amount rejected before persistence",
			ctx:    principalCtxFor(txTestBuyerID),
			reason: "bad match",
			prepare: func(fx *txFixture) {
				fx.matchRepo.stored.MatchedAmount = 0
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:   "offer already active is not updated",
			ctx:    principalCtxFor(txTestBuyerID),
			reason: "no offer change",
			prepare: func(fx *txFixture) {
				offer := txMustOffer()
				offer.Status = domain.OfferActive
				fx.offerRepo.stored = *offer
			},
			wantCascade:    true,
			offerUntouched: true,
		},
		{
			name:      "transaction update error stops cascade",
			ctx:       principalCtxFor(txTestBuyerID),
			reason:    "repo down",
			updateErr: errFake,
			wantErr:   errFake,
		},
		{
			name:           "match update error stops cascade",
			ctx:            principalCtxFor(txTestBuyerID),
			reason:         "repo down",
			matchUpdateErr: errFake,
			wantErr:        errFake,
		},
		{
			name:             "request update error stops cascade",
			ctx:              principalCtxFor(txTestBuyerID),
			reason:           "repo down",
			requestUpdateErr: errFake,
			wantErr:          errFake,
		},
		{
			name:           "offer update error propagated",
			ctx:            principalCtxFor(txTestBuyerID),
			reason:         "repo down",
			offerUpdateErr: errFake,
			wantErr:        errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			if tt.prepare != nil {
				tt.prepare(fx)
			}
			if tt.getErr != nil {
				fx.transactionRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
					return domain.Transaction{}, tt.getErr
				}
			}
			if tt.matchErr != nil {
				fx.matchRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
					return nil, tt.matchErr
				}
			}
			if tt.updateErr != nil {
				fx.transactionRepo.update = func(ctx context.Context, transaction *domain.Transaction) error {
					return tt.updateErr
				}
			}
			if tt.matchUpdateErr != nil {
				fx.matchRepo.update = func(ctx context.Context, match *domain.Match) error {
					return tt.matchUpdateErr
				}
			}
			if tt.requestUpdateErr != nil {
				fx.requestRepo.update = func(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
					return tt.requestUpdateErr
				}
			}
			if tt.offerUpdateErr != nil {
				fx.offerRepo.update = func(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
					return tt.offerUpdateErr
				}
			}

			err := fx.uc.Cancel(tt.ctx, txTestTransactID, tt.reason)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				repoFailure := tt.updateErr != nil || tt.matchUpdateErr != nil || tt.requestUpdateErr != nil || tt.offerUpdateErr != nil
				if !repoFailure && len(fx.transactionRepo.updated) != 0 {
					t.Fatalf("expected transaction not persisted, got %d updates", len(fx.transactionRepo.updated))
				}
				if tt.matchUpdateErr != nil && len(fx.matchRepo.updated) != 0 {
					t.Fatal("expected match update to fail before persistence")
				}
				if tt.requestUpdateErr != nil && len(fx.requestRepo.updated) != 0 {
					t.Fatal("expected request update to fail before persistence")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fx.transactionRepo.updated) != 1 {
				t.Fatalf("expected transaction persisted once, got %d", len(fx.transactionRepo.updated))
			}
			persisted := fx.transactionRepo.updated[0]
			if persisted.Status != domain.TransactionCancelled {
				t.Errorf("transaction status = %v, want cancelled", persisted.Status)
			}
			if persisted.CancelReason != tt.reason {
				t.Errorf("cancel reason = %q, want %q", persisted.CancelReason, tt.reason)
			}
			if tt.wantCancelledBy != nil {
				if persisted.CancelledBy == nil || *persisted.CancelledBy != *tt.wantCancelledBy {
					t.Errorf("cancelled by = %v, want %v", persisted.CancelledBy, *tt.wantCancelledBy)
				}
			}
			if !tt.wantCascade {
				if len(fx.matchRepo.updated) != 0 {
					t.Errorf("expected no match update, got %d", len(fx.matchRepo.updated))
				}
				if len(fx.offerRepo.updated) != 0 {
					t.Errorf("expected no offer update, got %d", len(fx.offerRepo.updated))
				}
				if len(fx.requestRepo.updated) != 0 {
					t.Errorf("expected no request update, got %d", len(fx.requestRepo.updated))
				}
				return
			}
			if len(fx.matchRepo.updated) != 1 || fx.matchRepo.updated[0].Status != domain.MatchCancelled {
				t.Errorf("expected match cancelled, got %+v", fx.matchRepo.updated)
			}
			if len(fx.requestRepo.updated) != 1 {
				t.Fatalf("expected request persisted once, got %d", len(fx.requestRepo.updated))
			}
			if got := fx.requestRepo.updated[0].ActualAmount; got != 100 {
				t.Errorf("released actual amount = %v, want 100", got)
			}
			if tt.offerUntouched {
				if len(fx.offerRepo.updated) != 0 {
					t.Errorf("expected no offer update, got %d", len(fx.offerRepo.updated))
				}
				return
			}
			if len(fx.offerRepo.updated) != 1 || fx.offerRepo.updated[0].Status != domain.OfferActive {
				t.Errorf("expected offer reactivated, got %+v", fx.offerRepo.updated)
			}
		})
	}
}

func TestTransactionUseCaseRejectsInvalidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  domain.TransactionStatus
		op      string
		wantErr error
	}{
		{name: "confirm start from in progress", status: domain.TransactionInProgress, op: "start", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "confirm start from completed", status: domain.TransactionCompleted, op: "start", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "confirm start from cancelled", status: domain.TransactionCancelled, op: "start", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "confirm delivery from matched", status: domain.TransactionMatched, op: "delivery", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "confirm delivery from completed", status: domain.TransactionCompleted, op: "delivery", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "confirm delivery from cancelled", status: domain.TransactionCancelled, op: "delivery", wantErr: domain.ErrInvalidTransactionTransition},
		{name: "cancel from completed", status: domain.TransactionCompleted, op: "cancel", wantErr: domain.ErrTerminalState},
		{name: "cancel from cancelled", status: domain.TransactionCancelled, op: "cancel", wantErr: domain.ErrTerminalState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			transaction := txInProgressTransaction()
			transaction.Status = tt.status
			if tt.status == domain.TransactionMatched {
				transaction = *txMustTransaction()
			}
			if tt.status == domain.TransactionCancelled {
				transaction.Status = domain.TransactionCancelled
				transaction.BuyerStartConfirmedAt = nil
				transaction.SupplierStartConfirmedAt = nil
				reason := "already cancelled"
				transaction.CancelReason = reason
			}
			fx.transactionRepo.stored = transaction

			ctx := principalCtxFor(txTestBuyerID)
			var err error
			switch tt.op {
			case "start":
				err = fx.uc.ConfirmStart(ctx, txTestTransactID)
			case "delivery":
				err = fx.uc.ConfirmDelivery(ctx, txTestTransactID)
			case "cancel":
				err = fx.uc.Cancel(ctx, txTestTransactID, "cancelling")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if len(fx.transactionRepo.updated) != 0 {
				t.Fatalf("expected no persistence, got %d updates", len(fx.transactionRepo.updated))
			}
			if len(fx.matchRepo.updated) != 0 || len(fx.requestRepo.updated) != 0 || len(fx.offerRepo.updated) != 0 {
				t.Fatal("expected no cascade persistence on rejected transition")
			}
		})
	}
}

func TestTransactionDTOHistory(t *testing.T) {
	t.Parallel()

	buyerStart := fixedTime.Add(time.Minute)
	supplierStart := fixedTime.Add(4 * time.Minute)
	buyerDelivery := fixedTime.Add(6 * time.Minute)
	supplierDelivery := fixedTime.Add(5 * time.Minute)
	cancelledAt := fixedTime.Add(10 * time.Minute)
	cancelledBy := txTestBuyerID

	tests := []struct {
		name      string
		prepare   func(transaction *domain.Transaction)
		wantLen   int
		wantTypes []domain.TransactionStatus
		wantAts   []time.Time
	}{
		{
			name:      "fresh transaction only has matched",
			prepare:   func(transaction *domain.Transaction) {},
			wantLen:   1,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched},
			wantAts:   []time.Time{fixedTime},
		},
		{
			name: "single start confirmation adds no state entry",
			prepare: func(transaction *domain.Transaction) {
				transaction.BuyerStartConfirmedAt = &buyerStart
			},
			wantLen:   1,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched},
			wantAts:   []time.Time{fixedTime},
		},
		{
			name: "both start confirmations add in progress at later timestamp",
			prepare: func(transaction *domain.Transaction) {
				transaction.BuyerStartConfirmedAt = &buyerStart
				transaction.SupplierStartConfirmedAt = &supplierStart
				transaction.Status = domain.TransactionInProgress
			},
			wantLen:   2,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched, domain.TransactionInProgress},
			wantAts:   []time.Time{fixedTime, supplierStart},
		},
		{
			name: "completed transaction adds completed at later delivery timestamp",
			prepare: func(transaction *domain.Transaction) {
				transaction.BuyerStartConfirmedAt = &buyerStart
				transaction.SupplierStartConfirmedAt = &supplierStart
				transaction.BuyerDeliveryConfirmedAt = &buyerDelivery
				transaction.SupplierDeliveryConfirmedAt = &supplierDelivery
				transaction.Status = domain.TransactionCompleted
				transaction.UpdatedAt = buyerDelivery
			},
			wantLen:   3,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched, domain.TransactionInProgress, domain.TransactionCompleted},
			wantAts:   []time.Time{fixedTime, supplierStart, buyerDelivery},
		},
		{
			name: "cancelled from in progress adds cancelled with reason",
			prepare: func(transaction *domain.Transaction) {
				transaction.BuyerStartConfirmedAt = &buyerStart
				transaction.SupplierStartConfirmedAt = &supplierStart
				transaction.Status = domain.TransactionCancelled
				transaction.UpdatedAt = cancelledAt
				transaction.CancelledBy = &cancelledBy
				transaction.CancelReason = "buyer changed plans"
			},
			wantLen:   3,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched, domain.TransactionInProgress, domain.TransactionCancelled},
			wantAts:   []time.Time{fixedTime, supplierStart, cancelledAt},
		},
		{
			name: "cancelled from matched adds cancelled after matched",
			prepare: func(transaction *domain.Transaction) {
				transaction.Status = domain.TransactionCancelled
				transaction.UpdatedAt = cancelledAt
				transaction.CancelledBy = &cancelledBy
				transaction.CancelReason = "no supplier available"
			},
			wantLen:   2,
			wantTypes: []domain.TransactionStatus{domain.TransactionMatched, domain.TransactionCancelled},
			wantAts:   []time.Time{fixedTime, cancelledAt},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newTxFixture()
			transaction := txMustTransaction()
			tt.prepare(transaction)
			fx.transactionRepo.stored = *transaction

			got, err := fx.uc.GetByMatch(principalCtxFor(txTestBuyerID), txTestMatchID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got.History) != tt.wantLen {
				t.Fatalf("history len = %d, want %d (%+v)", len(got.History), tt.wantLen, got.History)
			}
			for i, wantStatus := range tt.wantTypes {
				if got.History[i].Status != wantStatus {
					t.Errorf("history[%d].status = %v, want %v", i, got.History[i].Status, wantStatus)
				}
				if !got.History[i].At.Equal(tt.wantAts[i]) {
					t.Errorf("history[%d].at = %v, want %v", i, got.History[i].At, tt.wantAts[i])
				}
			}
			if tt.wantLen > 0 && tt.wantTypes[tt.wantLen-1] == domain.TransactionCancelled {
				last := got.History[tt.wantLen-1]
				if last.CancelReason == "" {
					t.Error("expected cancel reason on cancelled history entry")
				}
				if last.CancelledBy == nil || *last.CancelledBy != cancelledBy {
					t.Errorf("cancelled by = %v, want %v", last.CancelledBy, cancelledBy)
				}
			}
		})
	}
}
