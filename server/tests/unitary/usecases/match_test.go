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

type matchFixture struct {
	requests *fakeMatchSupplyRequestRepo
	offers   *fakeMatchSupplyOfferRepo
	matches  *fakeMatchRepository
	txs      *fakeMatchTransactionRepo
	recs     *stubMatchRecommendationUC
	uow      *fakeUnitOfWork
	uc       *usecases.MatchUseCaseImpl
}

func newMatchFixture() *matchFixture {
	f := &matchFixture{
		requests: newFakeMatchSupplyRequestRepo(),
		offers:   newFakeMatchSupplyOfferRepo(),
		matches:  newFakeMatchRepository(),
		txs:      newFakeMatchTransactionRepo(),
		recs:     &stubMatchRecommendationUC{},
	}
	f.uow = newFakeUnitOfWork(f.newTxScope())
	f.uc = usecases.NewMatchUseCase(f.requests, f.offers, f.matches, f.txs, f.recs, f.uow)
	return f
}

func (f *matchFixture) assertNoSideEffects(t *testing.T) {
	t.Helper()
	if len(f.requests.updated) != 0 {
		t.Errorf("request updates = %d, want 0", len(f.requests.updated))
	}
	if len(f.matches.created) != 0 {
		t.Errorf("matches created = %d, want 0", len(f.matches.created))
	}
	if len(f.txs.created) != 0 {
		t.Errorf("transactions created = %d, want 0", len(f.txs.created))
	}
	if len(f.offers.updated) != 0 {
		t.Errorf("offer updates = %d, want 0", len(f.offers.updated))
	}
}

func (f *matchFixture) storedRequest() domain.SupplyRequest {
	return f.requests.store[matchTestRequestID]
}

func TestMatchUseCaseLikeRejectsInvalidStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		setup      func(f *matchFixture)
		wantErr    error
		wantNoSide bool
	}{
		{
			name:       "unauthenticated",
			ctx:        context.Background(),
			wantErr:    auth.ErrUnauthenticated,
			wantNoSide: true,
		},
		{
			name: "offer not found",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.offers.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
					return domain.SupplyOffer{}, domain.ErrNotFound
				}
			},
			wantErr:    domain.ErrNotFound,
			wantNoSide: true,
		},
		{
			name: "offer already matched",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				offer := matchTestOffer()
				offer.Status = domain.OfferMatched
				f.offers.seed(*offer)
			},
			wantErr:    domain.ErrInvalidOfferStatus,
			wantNoSide: true,
		},
		{
			name: "offer already passed",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				offer := matchTestOffer()
				offer.Status = domain.OfferRejected
				f.offers.seed(*offer)
			},
			wantErr:    domain.ErrInvalidOfferStatus,
			wantNoSide: true,
		},
		{
			name: "offer withdrawn",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				offer := matchTestOffer()
				offer.Status = domain.OfferWithdrawn
				f.offers.seed(*offer)
			},
			wantErr:    domain.ErrInvalidOfferStatus,
			wantNoSide: true,
		},
		{
			name: "request not found",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.requests.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
					return domain.SupplyRequest{}, domain.ErrNotFound
				}
			},
			wantErr:    domain.ErrNotFound,
			wantNoSide: true,
		},
		{
			name: "request not open",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				request := matchTestRequest()
				request.Status = domain.SupplyRequestCompleted
				f.requests.store[matchTestRequestID] = *request
			},
			wantErr:    domain.ErrInvalidRequestStatus,
			wantNoSide: true,
		},
		{
			name:       "caller is not the buyer",
			ctx:        principalCtxFor(matchTestSupplierID),
			wantErr:    domain.ErrForbidden,
			wantNoSide: true,
		},
		{
			name: "offer already has an active match",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.matches.existsActiveByOffer = func(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			wantErr:    domain.ErrInvalidMatchStatus,
			wantNoSide: true,
		},
		{
			name: "single provider request already matched",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.requests.store[matchTestRequestID] = *matchTestSingleProviderRequest()
				f.matches.existsActiveByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			wantErr:    domain.ErrInvalidMatchStatus,
			wantNoSide: true,
		},
		{
			name: "offer amount exceeds remaining availability",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				request := matchTestRequest()
				request.ActualAmount = 10
				f.requests.store[matchTestRequestID] = *request
			},
			wantErr:    domain.ErrInsufficientAmount,
			wantNoSide: true,
		},
		{
			name: "existence check fails",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.matches.existsActiveByOffer = func(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
					return false, errFake
				}
			},
			wantErr:    errFake,
			wantNoSide: true,
		},
		{
			name: "request persistence fails",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.requests.update = func(ctx context.Context, request *domain.SupplyRequest) error {
					return errFake
				}
			},
			wantErr:    errFake,
			wantNoSide: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := newMatchFixture()
			if tt.setup != nil {
				tt.setup(f)
			}

			match, transaction, err := f.uc.Like(tt.ctx, matchTestOfferID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if match != nil || transaction != nil {
				t.Errorf("expected nil match and transaction on error, got %v / %v", match, transaction)
			}
			if tt.wantNoSide {
				f.assertNoSideEffects(t)
			}
		})
	}
}

func TestMatchUseCaseLikeHappyPathMultipleProviders(t *testing.T) {
	t.Parallel()

	f := newMatchFixture()

	match, transaction, err := f.uc.Like(principalCtx(), matchTestOfferID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.matches.created) != 1 {
		t.Fatalf("matches created = %d, want 1", len(f.matches.created))
	}
	created := f.matches.created[0]
	if created.MatchedAmount != 30 {
		t.Errorf("matched amount = %v, want 30", created.MatchedAmount)
	}
	if created.SupplyOffer != matchTestOfferID {
		t.Errorf("match offer = %v, want %v", created.SupplyOffer, matchTestOfferID)
	}
	if created.SupplyRequest != matchTestRequestID {
		t.Errorf("match request = %v, want %v", created.SupplyRequest, matchTestRequestID)
	}
	if !created.IsActive() {
		t.Errorf("match status = %v, want active", created.Status)
	}
	if created.AmountUnit != domain.Kg {
		t.Errorf("amount unit = %v, want kg", created.AmountUnit)
	}

	if len(f.txs.created) != 1 {
		t.Fatalf("transactions created = %d, want 1", len(f.txs.created))
	}
	createdTx := f.txs.created[0]
	if createdTx.MatchID != created.ID {
		t.Errorf("transaction match id = %v, want %v", createdTx.MatchID, created.ID)
	}
	if createdTx.Status != domain.TransactionMatched {
		t.Errorf("transaction status = %v, want matched", createdTx.Status)
	}

	if match.ID != created.ID {
		t.Errorf("dto match id = %v, want %v", match.ID, created.ID)
	}
	if transaction.ID == nil || *transaction.ID != createdTx.ID {
		t.Errorf("dto transaction id = %v, want %v", transaction.ID, createdTx.ID)
	}
	if transaction.MatchID == nil || *transaction.MatchID != created.ID {
		t.Errorf("dto transaction match id = %v, want %v", transaction.MatchID, created.ID)
	}

	if got := f.storedRequest().ActualAmount; got != 70 {
		t.Errorf("actual amount = %v, want 70", got)
	}
	if len(f.requests.updated) != 1 {
		t.Errorf("request updates = %d, want 1", len(f.requests.updated))
	}

	if got := f.offers.store[matchTestOfferID].Status; got != domain.OfferMatched {
		t.Errorf("offer status = %v, want matched", got)
	}
	if len(f.offers.updated) != 1 {
		t.Errorf("offer updates = %d, want 1", len(f.offers.updated))
	}
}

func TestMatchUseCaseLikeSingleProviderRejectsOtherActiveOffers(t *testing.T) {
	t.Parallel()

	f := newMatchFixture()
	f.requests.store[matchTestRequestID] = *matchTestSingleProviderRequest()

	other := matchTestOffer()
	other.ID = matchTestOtherOfferID
	passed := matchTestOffer()
	passed.ID = matchTestThirdOfferID
	passed.Status = domain.OfferRejected
	f.offers.seed(*other, *passed)

	_, _, err := f.uc.Like(principalCtx(), matchTestOfferID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := f.offers.store[matchTestOtherOfferID].Status; got != domain.OfferRejected {
		t.Errorf("other offer status = %v, want rejected", got)
	}
	if got := f.offers.store[matchTestThirdOfferID].Status; got != domain.OfferRejected {
		t.Errorf("passed offer status = %v, want unchanged rejected", got)
	}
	if got := f.offers.store[matchTestOfferID].Status; got != domain.OfferMatched {
		t.Errorf("liked offer status = %v, want matched", got)
	}
	if len(f.offers.updated) != 2 {
		t.Errorf("offer updates = %d, want 2 (liked + rejected)", len(f.offers.updated))
	}
	if len(f.matches.created) != 1 || len(f.txs.created) != 1 {
		t.Errorf("created match/tx = %d/%d, want 1/1", len(f.matches.created), len(f.txs.created))
	}
}

func TestMatchUseCasePass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(f *matchFixture)
		wantErr error
	}{
		{
			name:    "unauthenticated",
			ctx:     context.Background(),
			wantErr: auth.ErrUnauthenticated,
		},
		{
			name: "offer not found",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.offers.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
					return domain.SupplyOffer{}, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "offer already matched",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				offer := matchTestOffer()
				offer.Status = domain.OfferMatched
				f.offers.seed(*offer)
			},
			wantErr: domain.ErrInvalidOfferStatus,
		},
		{
			name:    "caller is not the buyer",
			ctx:     principalCtxFor(matchTestSupplierID),
			wantErr: domain.ErrForbidden,
		},
		{
			name: "persistence fails",
			ctx:  principalCtx(),
			setup: func(f *matchFixture) {
				f.offers.update = func(ctx context.Context, offer *domain.SupplyOffer) error {
					return errFake
				}
			},
			wantErr: errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := newMatchFixture()
			if tt.setup != nil {
				tt.setup(f)
			}

			err := f.uc.Pass(tt.ctx, matchTestOfferID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				return
			}
			if len(f.matches.created) != 0 {
				t.Errorf("matches created = %d, want 0", len(f.matches.created))
			}
			if len(f.offers.updated) != 0 {
				t.Errorf("offer updates = %d, want 0", len(f.offers.updated))
			}
		})
	}
}

func TestMatchUseCasePassRejectsOfferPermanently(t *testing.T) {
	t.Parallel()

	f := newMatchFixture()
	other := matchTestOffer()
	other.ID = matchTestOtherOfferID
	f.offers.seed(*other)

	if err := f.uc.Pass(principalCtx(), matchTestOfferID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := f.offers.store[matchTestOfferID].Status; got != domain.OfferRejected {
		t.Errorf("offer status = %v, want rejected", got)
	}
	if len(f.offers.updated) != 1 {
		t.Errorf("offer updates = %d, want 1", len(f.offers.updated))
	}
	if len(f.matches.created) != 0 {
		t.Errorf("matches created = %d, want 0", len(f.matches.created))
	}
	if got := f.offers.store[matchTestOtherOfferID].Status; got != domain.OfferActive {
		t.Errorf("other offer status = %v, want active", got)
	}

	_, _, err := f.uc.Like(principalCtx(), matchTestOfferID)
	if !errors.Is(err, domain.ErrInvalidOfferStatus) {
		t.Errorf("like after pass = %v, want %v", err, domain.ErrInvalidOfferStatus)
	}
}

func TestMatchUseCaseGetByID(t *testing.T) {
	t.Parallel()

	newFixtureWithMatch := func() *matchFixture {
		f := newMatchFixture()
		f.matches.getByID = func(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
			if matchID != matchTestMatchID {
				return nil, domain.ErrNotFound
			}
			return domain.NewMatch(matchTestOfferID, matchTestRequestID, 30, domain.Kg), nil
		}
		return f
	}

	t.Run("buyer can read the match", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()

		got, err := f.uc.GetByID(principalCtx(), matchTestMatchID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SupplyOffer != matchTestOfferID {
			t.Errorf("offer id = %v, want %v", got.SupplyOffer, matchTestOfferID)
		}
		if got.MatchedAmount != 30 {
			t.Errorf("matched amount = %v, want 30", got.MatchedAmount)
		}
	})

	t.Run("supplier can read the match", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()

		if _, err := f.uc.GetByID(principalCtxFor(matchTestSupplierID), matchTestMatchID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("third party is forbidden", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()

		_, err := f.uc.GetByID(principalCtxFor(uuid.New()), matchTestMatchID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()

		_, err := f.uc.GetByID(context.Background(), matchTestMatchID)
		if !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatalf("error = %v, want %v", err, auth.ErrUnauthenticated)
		}
	})

	t.Run("match not found", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()

		_, err := f.uc.GetByID(principalCtx(), uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("request load fails", func(t *testing.T) {
		t.Parallel()
		f := newFixtureWithMatch()
		f.requests.getByID = func(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
			return domain.SupplyRequest{}, errFake
		}

		_, err := f.uc.GetByID(principalCtx(), matchTestMatchID)
		if !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})
}

func TestMatchUseCaseListByRequest(t *testing.T) {
	t.Parallel()

	setupList := func(f *matchFixture) {
		first := domain.NewMatch(matchTestOfferID, matchTestRequestID, 30, domain.Kg)
		second := domain.NewMatch(matchTestOtherOfferID, matchTestRequestID, 20, domain.Kg)
		f.matches.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
			return []domain.Match{*first, *second}, nil
		}
	}

	t.Run("buyer lists matches", func(t *testing.T) {
		t.Parallel()
		f := newMatchFixture()
		setupList(f)

		got, err := f.uc.ListByRequest(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("matches = %d, want 2", len(got))
		}
		if got[0].MatchedAmount != 30 || got[1].MatchedAmount != 20 {
			t.Errorf("matched amounts = %v / %v, want 30 / 20", got[0].MatchedAmount, got[1].MatchedAmount)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		f := newMatchFixture()

		_, err := f.uc.ListByRequest(context.Background(), matchTestRequestID)
		if !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatalf("error = %v, want %v", err, auth.ErrUnauthenticated)
		}
	})

	t.Run("request not found", func(t *testing.T) {
		t.Parallel()
		f := newMatchFixture()

		_, err := f.uc.ListByRequest(principalCtx(), uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("non-buyer is forbidden", func(t *testing.T) {
		t.Parallel()
		f := newMatchFixture()

		_, err := f.uc.ListByRequest(principalCtxFor(matchTestSupplierID), matchTestRequestID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		f := newMatchFixture()
		f.matches.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
			return nil, errFake
		}

		_, err := f.uc.ListByRequest(principalCtx(), matchTestRequestID)
		if !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})
}

func TestMatchUseCaseListPrioritizedDelegatesToRecommendations(t *testing.T) {
	t.Parallel()

	f := newMatchFixture()
	ranked := []*dto.PrioritizedOfferDTO{{Score: 42}}
	f.recs.ranked = ranked

	got, err := f.uc.ListPrioritized(principalCtx(), matchTestRequestID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Score != 42 {
		t.Fatalf("result = %v, want the ranked list from the recommendation use case", got)
	}
	if f.recs.gotRequestID != matchTestRequestID {
		t.Errorf("request id passed = %v, want %v", f.recs.gotRequestID, matchTestRequestID)
	}

	f.recs.err = domain.ErrForbidden
	if _, err := f.uc.ListPrioritized(principalCtx(), matchTestRequestID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("error = %v, want %v", err, domain.ErrForbidden)
	}
}
