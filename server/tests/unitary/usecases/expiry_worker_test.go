package usecases_test

import (
	"context"
	"slices"
	"testing"
	"time"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

// TestExpiryWorkerSweepPurgesEveryCacheOfAnExpiredOffering pins the three
// evictions a deactivation owes: the search document, the single-offering
// cache entry that GET /offerings/{id} serves, and the owner's catalogue list.
// A missing offering:<id> purge leaves the API answering is_active:true for a
// product the database has already deactivated, for up to the five-minute TTL.
func TestExpiryWorkerSweepPurgesEveryCacheOfAnExpiredOffering(t *testing.T) {
	t.Parallel()

	offering := mustOffering()

	offeringRepo := newFakeOfferingRepo()
	offeringRepo.deactivateExpired = func(ctx context.Context, now time.Time) ([]domain.Offering, error) {
		return []domain.Offering{*offering}, nil
	}
	fuzzy := &fakeFuzzyRetrival{}
	cache := newFakeCache()
	invalidator := &fakeInvalidator{}

	worker := usecases.NewExpiryWorker(offeringRepo, fuzzy, cache, invalidator, newFakeTimer())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	worker.Run(ctx)

	if !slices.Contains(fuzzy.deletedIDs, offering.ID.String()) {
		t.Errorf("fuzzy deletes = %v, want the expired offering %s", fuzzy.deletedIDs, offering.ID)
	}
	if !slices.Contains(cache.deletedKeys, "offering:"+offering.ID.String()) {
		t.Errorf("cache deletes = %v, want the offering entry offering:%s", cache.deletedKeys, offering.ID)
	}
	if !slices.Contains(cache.deletedKeys, "offerings:byuser:"+offering.UserID.String()) {
		t.Errorf("cache deletes = %v, want the catalogue entry offerings:byuser:%s", cache.deletedKeys, offering.UserID)
	}
	if !invalidator.called {
		t.Error("the search cache was not invalidated after a real deactivation")
	}
}

// TestExpiryWorkerEmptySweepLeavesTheCachesAlone is the other half of the
// contract: a sweep that deactivated nothing must not evict anything, or every
// five-minute tick would cold-start every cached catalogue for no reason.
func TestExpiryWorkerEmptySweepLeavesTheCachesAlone(t *testing.T) {
	t.Parallel()

	offeringRepo := newFakeOfferingRepo()
	offeringRepo.deactivateExpired = func(ctx context.Context, now time.Time) ([]domain.Offering, error) {
		return nil, nil
	}
	fuzzy := &fakeFuzzyRetrival{}
	cache := newFakeCache()
	invalidator := &fakeInvalidator{}

	worker := usecases.NewExpiryWorker(offeringRepo, fuzzy, cache, invalidator, newFakeTimer())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	worker.Run(ctx)

	if len(fuzzy.deletedIDs) != 0 {
		t.Errorf("fuzzy deletes = %v, want none for an empty sweep", fuzzy.deletedIDs)
	}
	if len(cache.deletedKeys) != 0 {
		t.Errorf("cache deletes = %v, want none for an empty sweep", cache.deletedKeys)
	}
	if invalidator.called {
		t.Error("an empty sweep invalidated the search cache")
	}
}
