package usecases_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
)

func TestDistanceScoreFactor(t *testing.T) {
	t.Parallel()

	factor := usecases.DistanceScoreFactor()
	if factor.Name() != "distance" {
		t.Fatalf("name = %q, want distance", factor.Name())
	}

	ctx := context.Background()

	samePoint, err := factor.Score(ctx, primary.OfferScoreInput{DistanceKM: 0, HasDistance: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if samePoint != 1 {
		t.Errorf("same point score = %v, want 1", samePoint)
	}

	near, err := factor.Score(ctx, primary.OfferScoreInput{DistanceKM: 10, HasDistance: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	far, err := factor.Score(ctx, primary.OfferScoreInput{DistanceKM: 100, HasDistance: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(near-1.0/1.2) > 1e-9 {
		t.Errorf("near score = %v, want %v", near, 1.0/1.2)
	}
	if math.Abs(far-1.0/3.0) > 1e-9 {
		t.Errorf("far score = %v, want %v", far, 1.0/3.0)
	}
	if far >= near {
		t.Errorf("far score %v must be below near score %v", far, near)
	}

	unknown, err := factor.Score(ctx, primary.OfferScoreInput{HasDistance: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unknown != 0.5 {
		t.Errorf("unknown distance score = %v, want 0.5", unknown)
	}
}

func TestRecommendationDistanceChangesRanking(t *testing.T) {
	t.Parallel()

	buyer := domain.Address{Latitude: 12.1149926, Longitude: -86.2361744}
	near := domain.Address{Latitude: 12.1350, Longitude: -86.2500}
	far := domain.Address{Latitude: 12.4379, Longitude: -86.8781}

	f := newRecommendationFixture()
	request := matchTestRequest()
	request.Address = buyer
	f.requests.store[matchTestRequestID] = *request

	f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(time.Hour))
	f.seedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferActive, fixedTime)

	f.users.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		user := mustUser()
		user.ID = id
		if id == matchTestSupplierID {
			user.Address = near
		} else {
			user.Address = far
		}
		return user, nil
	}
	f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
		return matchTestInventory(100), nil
	}
	f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}

	tied, err := f.useCase([]usecases.WeightedScoreFactor{
		{Factor: usecases.AvailabilityScoreFactor(), Weight: 1},
	}).RankOffers(principalCtx(), matchTestRequestID)
	if err != nil {
		t.Fatalf("unexpected error without distance: %v", err)
	}
	if len(tied) != 2 {
		t.Fatalf("offers = %d, want 2", len(tied))
	}
	if tied[0].Offer.ID == nil || *tied[0].Offer.ID != matchTestOtherOfferID {
		t.Fatalf("without distance the older far offer should rank first, got %v", tied[0].Offer.ID)
	}

	ranked, err := f.useCase([]usecases.WeightedScoreFactor{
		{Factor: usecases.DistanceScoreFactor(), Weight: 1},
	}).RankOffers(principalCtx(), matchTestRequestID)
	if err != nil {
		t.Fatalf("unexpected error with distance: %v", err)
	}
	if len(ranked) != 2 {
		t.Fatalf("offers = %d, want 2", len(ranked))
	}
	if ranked[0].Offer.ID == nil || *ranked[0].Offer.ID != matchTestOfferID {
		t.Fatalf("with distance the nearer offer should rank first, got %v", ranked[0].Offer.ID)
	}
	if ranked[0].DistanceKM == nil || ranked[1].DistanceKM == nil {
		t.Fatal("distance_km must be populated for both ranked offers")
	}
	if *ranked[0].DistanceKM >= *ranked[1].DistanceKM {
		t.Errorf("nearer distance %v must be below farther distance %v", *ranked[0].DistanceKM, *ranked[1].DistanceKM)
	}
}

func TestRecommendationMissingSupplierRowKeepsOffersRanked(t *testing.T) {
	t.Parallel()

	f := newRecommendationFixture()
	request := matchTestRequest()
	request.Address = domain.Address{Latitude: 12.1149926, Longitude: -86.2361744}
	f.requests.store[matchTestRequestID] = *request

	f.seedOffer(matchTestOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(time.Hour))
	f.seedOffer(matchTestThirdOfferID, matchTestSupplierID, domain.OfferActive, fixedTime.Add(2*time.Hour))
	f.seedOffer(matchTestOtherOfferID, matchTestOtherSupply, domain.OfferActive, fixedTime)

	calls := map[uuid.UUID]int{}
	f.users.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		calls[id]++
		if id == matchTestSupplierID {
			return nil, domain.ErrNotFound
		}
		user := mustUser()
		user.ID = id
		user.Address = domain.Address{Latitude: 12.4379, Longitude: -86.8781}
		return user, nil
	}
	f.inventory.findBySupplierAndProduct = func(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
		return matchTestInventory(100), nil
	}
	f.matches.listActiveBySupplier = func(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
		return nil, nil
	}

	got, err := f.useCase(nil).RankOffers(principalCtx(), matchTestRequestID)
	if err != nil {
		t.Fatalf("a missing supplier row must not abort ranking: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("offers = %d, want 3", len(got))
	}

	seen := map[uuid.UUID]bool{}
	for _, ranked := range got {
		if ranked.Offer.ID == nil {
			t.Fatal("ranked offer has a nil id")
		}
		seen[*ranked.Offer.ID] = true
		if len(ranked.Contributions) != 5 {
			t.Fatalf("contributions = %d, want 5", len(ranked.Contributions))
		}
		distance := ranked.Contributions[0]
		if distance.Factor != "distance" {
			t.Fatalf("factor = %q, want distance", distance.Factor)
		}
		if *ranked.Offer.ID == matchTestOtherOfferID {
			if ranked.DistanceKM == nil {
				t.Error("a resolved supplier must expose distance_km")
			}
			continue
		}
		if ranked.DistanceKM != nil {
			t.Errorf("offer %v distance_km = %v, want nil", *ranked.Offer.ID, *ranked.DistanceKM)
		}
		if distance.Score != 0.5 {
			t.Errorf("offer %v distance score = %v, want neutral 0.5", *ranked.Offer.ID, distance.Score)
		}
	}
	for _, id := range []uuid.UUID{matchTestOfferID, matchTestOtherOfferID, matchTestThirdOfferID} {
		if !seen[id] {
			t.Errorf("offer %v missing from the ranked list", id)
		}
	}
	if calls[matchTestSupplierID] != 1 {
		t.Errorf("missing supplier resolved %d times, want 1", calls[matchTestSupplierID])
	}
}
