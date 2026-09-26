package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	"milpa/internal/auth"
)

type recommendationFixture struct {
	offers    *fakeMatchSupplyOfferRepo
	requests  *fakeMatchSupplyRequestRepo
	inventory *fakeMatchInventoryRepo
	matches   *fakeMatchRepository
}

func newRecommendationFixture() *recommendationFixture {
	return &recommendationFixture{
		offers:    newFakeMatchSupplyOfferRepo(),
		requests:  newFakeMatchSupplyRequestRepo(),
		inventory: newFakeMatchInventoryRepo(),
		matches:   newFakeMatchRepository(),
	}
}

func (f *recommendationFixture) useCase(factors []usecases.WeightedScoreFactor) *usecases.RecommendationUseCaseImpl {
	return usecases.NewRecommendationUseCase(f.offers, f.requests, f.inventory, f.matches, factors)
}

func (f *recommendationFixture) seedOffer(id, supplierID uuid.UUID, status domain.OfferStatus, createdAt time.Time) {
	offer := domain.NewSupplyOffer(supplierID, matchTestRequestID, 30, domain.Kg, fixedTime, true)
	offer.ID = id
	offer.Status = status
	offer.CreatedAt = createdAt
	f.offers.seed(*offer)
}

type stubScoreFactor struct {
	name  string
	score float64
	err   error
}

func (s stubScoreFactor) Name() string {
	return s.name
}

func (s stubScoreFactor) Score(ctx context.Context, input primary.OfferScoreInput) (float64, error) {
	return s.score, s.err
}

func TestRecommendationAvailableQuantity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(f *recommendationFixture)
		want    float32
		wantErr error
	}{
		{
			name: "inventory minus active matches",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					active := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
					return []domain.Match{*active}, nil
				}
			},
			want: 20,
		},
		{
			name: "no active matches returns full inventory",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					return nil, nil
				}
			},
			want: 100,
		},
		{
			name: "cancelled matches are not subtracted",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					cancelled := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
					cancelled.Status = domain.MatchCancelled
					return []domain.Match{*cancelled}, nil
				}
			},
			want: 100,
		},
		{
			name: "multiple active matches are summed",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					first := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
					second := domain.NewMatch(matchTestOtherOfferID, matchTestRequestID, 10, domain.Kg)
					return []domain.Match{*first, *second}, nil
				}
			},
			want: 10,
		},
		{
			name: "reserved amount above inventory clamps to zero",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					first := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
					second := domain.NewMatch(matchTestOtherOfferID, matchTestRequestID, 70, domain.Kg)
					return []domain.Match{*first, *second}, nil
				}
			},
			want: 0,
		},
		{
			name: "inventory not found",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
					return domain.SupplierInventory{}, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "inventory repository error",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
					return domain.SupplierInventory{}, errFake
				}
			},
			wantErr: errFake,
		},
		{
			name: "match repository error",
			ctx:  principalCtx(),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					return nil, errFake
				}
			},
			wantErr: errFake,
		},
		{
			name:    "unauthenticated",
			ctx:     context.Background(),
			wantErr: auth.ErrUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := newRecommendationFixture()
			if tt.setup != nil {
				tt.setup(f)
			}
			uc := f.useCase(nil)

			got, err := uc.AvailableQuantity(tt.ctx, matchTestSupplierID, matchTestProduct)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("available = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecommendationAvailableQuantityLooksUpInventory(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	uc := f.useCase(nil)

	if _, err := uc.AvailableQuantity(principalCtx(), matchTestSupplierID, matchTestProduct); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.inventory.findCalls) != 1 {
		t.Fatalf("inventory lookups = %d, want 1", len(f.inventory.findCalls))
	}
	want := matchTestSupplierID.String() + "/" + matchTestProduct
	if f.inventory.findCalls[0] != want {
		t.Errorf("inventory lookup = %q, want %q", f.inventory.findCalls[0], want)
	}
}

func TestRecommendationRankOffers(t *testing.T) {
	t.Parallel()

	t.Run("ranks active offers by availability descending", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(2*time.Hour))
		f.seedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferActive, fixedTime.Add(time.Hour))
		f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
			if supplierID == matchTestSupplierID {
				return matchTestInventory(100), nil
			}
			return *domain.NewSupplierInventory(supplierID, productName, 50, domain.Kg), nil
		}
		f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
			if supplierID == matchTestSupplierID {
				active := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
				return []domain.Match{*active}, nil
			}
			return nil, nil
		}

		uc := f.useCase(nil)
		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("offers = %d, want 2", len(got))
		}
		if got[0].Offer.ID == nil || *got[0].Offer.ID != matchTestOtherOfferID {
			t.Errorf("first offer = %v, want %v", got[0].Offer.ID, matchTestOtherOfferID)
		}
		if got[0].Score != 50 || got[0].AvailableQuantity != 50 {
			t.Errorf("first score/available = %v/%v, want 50/50", got[0].Score, got[0].AvailableQuantity)
		}
		if got[1].Score != 20 || got[1].AvailableQuantity != 20 {
			t.Errorf("second score/available = %v/%v, want 20/20", got[1].Score, got[1].AvailableQuantity)
		}
		if len(got[0].Contributions) != 1 {
			t.Fatalf("contributions = %d, want 1", len(got[0].Contributions))
		}
		contribution := got[0].Contributions[0]
		if contribution.Factor != "availability" {
			t.Errorf("factor = %q, want availability", contribution.Factor)
		}
		if contribution.Weight != 1 {
			t.Errorf("weight = %v, want 1", contribution.Weight)
		}
		if contribution.Score != 50 || contribution.WeightedScore != 50 {
			t.Errorf("contribution score/weighted = %v/%v, want 50/50", contribution.Score, contribution.WeightedScore)
		}
	})

	t.Run("excludes offers that are not actionable", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferMatched, fixedTime)
		f.seedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferRejected, fixedTime)
		f.seedOffer(matchTestThirdOfferID, matchTestSupplierID, domain.OfferWithdrawn, fixedTime)

		uc := f.useCase(nil)
		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("offers = %d, want 0", len(got))
		}
	})

	t.Run("missing inventory ranks with zero availability", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime)
		f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
			return domain.SupplierInventory{}, domain.ErrNotFound
		}

		uc := f.useCase(nil)
		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("offers = %d, want 1", len(got))
		}
		if got[0].Score != 0 || got[0].AvailableQuantity != 0 {
			t.Errorf("score/available = %v/%v, want 0/0", got[0].Score, got[0].AvailableQuantity)
		}
	})

	t.Run("ties break by oldest offer first", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(time.Hour))
		f.seedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferActive, fixedTime)

		uc := f.useCase(nil)
		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("offers = %d, want 2", len(got))
		}
		if got[0].Offer.ID == nil || *got[0].Offer.ID != matchTestOtherOfferID {
			t.Errorf("first offer = %v, want older offer %v", got[0].Offer.ID, matchTestOtherOfferID)
		}
	})

	t.Run("weighted factors compose without changing ranking code", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime)

		factors := []usecases.WeightedScoreFactor{
			{Factor: usecases.AvailabilityScoreFactor(), Weight: 1},
			{Factor: stubScoreFactor{name: "geography", score: 10}, Weight: 0.5},
		}
		uc := f.useCase(factors)

		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("offers = %d, want 1", len(got))
		}
		if got[0].Score != 105 {
			t.Errorf("score = %v, want 105 (100 availability + 0.5*10 geography)", got[0].Score)
		}
		if len(got[0].Contributions) != 2 {
			t.Fatalf("contributions = %d, want 2", len(got[0].Contributions))
		}
		if got[0].Contributions[1].Factor != "geography" || got[0].Contributions[1].WeightedScore != 5 {
			t.Errorf("second contribution = %+v, want geography weighted 5", got[0].Contributions[1])
		}
	})

	t.Run("factor error is propagated", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime)

		factors := []usecases.WeightedScoreFactor{
			{Factor: stubScoreFactor{name: "failing", err: errFake}, Weight: 1},
		}
		uc := f.useCase(factors)

		if _, err := uc.RankOffers(principalCtx(), matchTestRequestID); !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		uc := f.useCase(nil)

		if _, err := uc.RankOffers(context.Background(), matchTestRequestID); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatalf("error = %v, want %v", err, auth.ErrUnauthenticated)
		}
	})

	t.Run("request not found", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		uc := f.useCase(nil)

		if _, err := uc.RankOffers(principalCtx(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("non-buyer is forbidden", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		uc := f.useCase(nil)

		if _, err := uc.RankOffers(principalCtxFor(matchTestSupplierID), matchTestRequestID); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
		}
	})

	t.Run("offer listing error is propagated", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.offers.listByRequest = func(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
			return nil, errFake
		}
		uc := f.useCase(nil)

		if _, err := uc.RankOffers(principalCtx(), matchTestRequestID); !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})

	t.Run("inventory repository failure aborts ranking", func(t *testing.T) {
		t.Parallel()

		f := newRecommendationFixture()
		f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime)
		f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
			return domain.SupplierInventory{}, errFake
		}
		uc := f.useCase(nil)

		if _, err := uc.RankOffers(principalCtx(), matchTestRequestID); !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})
}
