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

// seedPricedOffer is seedOffer plus the RF-11 quotation. A nil price is a
// legitimate state for a row that predates the column, so the helper takes the
// pointer rather than hiding the nil case behind a default.
func (f *recommendationFixture) seedPricedOffer(id, supplierID uuid.UUID, status domain.OfferStatus, createdAt time.Time, price *float64, comments string) {
	offer := domain.NewSupplyOffer(supplierID, matchTestRequestID, 30, domain.Kg, fixedTime, true)
	offer.ID = id
	offer.Status = status
	offer.CreatedAt = createdAt
	offer.PricePerUnit = price
	offer.Comments = comments
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
		want    float64
		wantErr error
	}{
		{
			name: "inventory minus active matches",
			ctx:  principalCtxFor(matchTestSupplierID),
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
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					return nil, nil
				}
			},
			want: 100,
		},
		{
			name: "cancelled matches are not subtracted",
			ctx:  principalCtxFor(matchTestSupplierID),
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
			name: "active match on another product is not subtracted",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				otherProductRequest := domain.NewSupplyRequest(
					testUserID, "Beans", 100, domain.Kg, 1, 100, domain.Kg,
					domain.Address{}, fixedTime, fixedTime.Add(72*time.Hour), "", true,
				)
				f.requests.store[otherProductRequest.ID] = *otherProductRequest
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					active := domain.NewMatch(matchTestOfferID, otherProductRequest.ID, 80, domain.Kg)
					return []domain.Match{*active}, nil
				}
			},
			want: 100,
		},
		{
			name: "active match in another unit is not subtracted",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					active := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Lb)
					return []domain.Match{*active}, nil
				}
			},
			want: 100,
		},
		{
			name: "active match on same product and unit is subtracted",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					active := domain.NewMatch(matchTestOfferID, matchTestRequestID, 80, domain.Kg)
					return []domain.Match{*active}, nil
				}
			},
			want: 20,
		},
		{
			name: "active match on missing supply request is skipped",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					active := domain.NewMatch(matchTestOfferID, uuid.New(), 80, domain.Kg)
					return []domain.Match{*active}, nil
				}
			},
			want: 100,
		},
		{
			name: "multiple active matches are summed",
			ctx:  principalCtxFor(matchTestSupplierID),
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
			ctx:  principalCtxFor(matchTestSupplierID),
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
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
					return domain.SupplierInventory{}, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "inventory repository error",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
					return domain.SupplierInventory{}, errFake
				}
			},
			wantErr: errFake,
		},
		{
			name: "match repository error",
			ctx:  principalCtxFor(matchTestSupplierID),
			setup: func(f *recommendationFixture) {
				f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
					return nil, errFake
				}
			},
			wantErr: errFake,
		},
		{
			name:    "another supplier's availability is forbidden",
			ctx:     principalCtx(),
			wantErr: domain.ErrForbidden,
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

// A supplier reads their own availability.
func TestRecommendationAvailableQuantityAllowsOwnSupplier(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
		return matchTestInventory(100), nil
	}
	f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	uc := f.useCase(nil)

	got, err := uc.AvailableQuantity(principalCtxFor(matchTestSupplierID), matchTestSupplierID, matchTestProduct)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("available = %v, want 100", got)
	}
}

// A caller that is not the supplier is refused BEFORE any repository is read, so
// availability cannot be used as a side channel to probe another supplier's
// stock.
func TestRecommendationAvailableQuantityForbiddenWithoutRepositoryRead(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	uc := f.useCase(nil)

	got, err := uc.AvailableQuantity(principalCtx(), matchTestSupplierID, matchTestProduct)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
	}
	if got != 0 {
		t.Errorf("available = %v, want 0", got)
	}
	if len(f.inventory.findCalls) != 0 {
		t.Errorf("inventory lookups = %d, want 0", len(f.inventory.findCalls))
	}
}

// The ownership gate belongs to the PUBLIC read only. RankOffers resolves
// availability on behalf of the BUYER of a request, who is entitled to see a
// candidate supplier's stock. Gating the shared helper would break
// recommendation; this test pins that it still works.
func TestRecommendationRankOffersStillReadsOtherSuppliersAvailability(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime)
	f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
		if supplierID != matchTestSupplierID {
			t.Errorf("inventory resolved for %v, want %v", supplierID, matchTestSupplierID)
		}
		return matchTestInventory(100), nil
	}
	f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}
	uc := f.useCase(nil)

	// The principal here is the BUYER, not matchTestSupplierID.
	got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
	if err != nil {
		t.Fatalf("buyer ranking another supplier's offer must not be forbidden: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("offers = %d, want 1", len(got))
	}
	if got[0].AvailableQuantity != 100 {
		t.Errorf("available = %v, want 100", got[0].AvailableQuantity)
	}
}

func TestRecommendationAvailableQuantityLooksUpInventory(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	uc := f.useCase(nil)

	if _, err := uc.AvailableQuantity(principalCtxFor(matchTestSupplierID), matchTestSupplierID, matchTestProduct); err != nil {
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

	t.Run("carries the price and comments to the prioritized payload", func(t *testing.T) {
		t.Parallel()

		// The regression test for the duplicated mapper: recommendation.go used
		// to carry its own matchOfferToDTO, so every field added to the offer DTO
		// had to be remembered in two places, and forgetting one failed silently
		// on the one endpoint where a buyer compares suppliers. So the assertion
		// is on the ranked payload, not on the DTO.
		f := newRecommendationFixture()
		f.seedPricedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(time.Hour), ptrFloat64(8.25), "harvested last week")
		f.seedPricedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferActive, fixedTime, nil, "quote pending")

		uc := f.useCase(nil)
		got, err := uc.RankOffers(principalCtx(), matchTestRequestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("offers = %d, want 2", len(got))
		}

		byID := make(map[uuid.UUID]struct {
			price    *float64
			comments string
		}, len(got))
		for _, ranked := range got {
			if ranked.Offer.ID == nil {
				t.Fatal("ranked offer has a nil id")
			}
			byID[*ranked.Offer.ID] = struct {
				price    *float64
				comments string
			}{ranked.Offer.PricePerUnit, ranked.Offer.Comments}
		}

		priced, ok := byID[matchTestOfferID]
		if !ok {
			t.Fatalf("the priced offer %s is missing from the ranked payload", matchTestOfferID)
		}
		if priced.price == nil {
			t.Fatal("ranked offer price_per_unit = null, want 8.25")
		}
		if *priced.price != 8.25 {
			t.Errorf("ranked offer price_per_unit = %v, want 8.25", *priced.price)
		}
		if priced.comments != "harvested last week" {
			t.Errorf("ranked offer comments = %q, want %q", priced.comments, "harvested last week")
		}

		// The other direction matters just as much: an offer that predates the
		// price column has to reach the buyer as null, not as a zero. A zero
		// here would read as the cheapest offer on the platform.
		unpriced, ok := byID[matchTestOtherOfferID]
		if !ok {
			t.Fatalf("the unpriced offer %s is missing from the ranked payload", matchTestOtherOfferID)
		}
		if unpriced.price != nil {
			t.Errorf("ranked offer price_per_unit = %v, want null for an offer with no quote", *unpriced.price)
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
