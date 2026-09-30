package usecases_test

import (
	"errors"
	"testing"
	"time"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

func cachedOfferingUC(c port.Cache) *usecases.CachedOfferingUseCase {
	return usecases.NewCachedOfferingUseCase(
		usecases.NewOfferingUseCase(
			newFakeOfferingRepo(),
			newFakeUserRepo(),
			newFakeTimer(),
			&fakeFuzzyRetrival{},
			&fakeInvalidator{},
		),
		c,
	)
}

func storingCacheKeys(c *storingCache) map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make(map[string]bool, len(c.items))
	for key := range c.items {
		keys[key] = true
	}
	return keys
}

func warmOfferingCache(t *testing.T, uc *usecases.CachedOfferingUseCase) {
	t.Helper()

	if _, err := uc.GetByID(principalCtx(), testOfferingID); err != nil {
		t.Fatalf("warm the offering entry: %v", err)
	}
	if _, err := uc.GetByUserID(principalCtx(), testUserID); err != nil {
		t.Fatalf("warm the farmer list entry: %v", err)
	}
}

func TestCachedOfferingUpdateDropsTheFarmerList(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedOfferingUC(cache)

	warmOfferingCache(t, uc)

	warmed := storingCacheKeys(cache)
	if !warmed["offering:"+testOfferingID.String()] {
		t.Fatalf("the single-offering key was never cached: %v", warmed)
	}
	if !warmed["offerings:byuser:"+testUserID.String()] {
		t.Fatalf("the farmer list key was never cached: %v", warmed)
	}

	if err := uc.UpdateOffering(principalCtx(), testOfferingID, dto.UpdateOfferingRequest{
		Name:              strPtr("Maiz dulce"),
		CategoryID:        &testCategoryID,
		QuantityAvailable: floatPtr(80),
	}); err != nil {
		t.Fatalf("UpdateOffering() error: %v", err)
	}

	remaining := storingCacheKeys(cache)
	if remaining["offering:"+testOfferingID.String()] {
		t.Error("the single-offering key survived an edit")
	}
	if remaining["offerings:byuser:"+testUserID.String()] {
		t.Errorf("the farmer list key survived an edit: %v", remaining)
	}
}

func TestCachedOfferingRenewDropsTheOfferingAndTheFarmerList(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedOfferingUC(cache)

	warmOfferingCache(t, uc)

	newExpiry := fixedTime.Add(15 * 24 * time.Hour)

	renewed, err := uc.RenewOffering(principalCtx(), testOfferingID, dto.RenewOfferingRequest{ExpiresAt: newExpiry})
	if err != nil {
		t.Fatalf("RenewOffering() error: %v", err)
	}
	if !renewed.IsActive {
		t.Error("the renewed offering is not active")
	}
	if renewed.ExpiresAt == nil || !renewed.ExpiresAt.Equal(newExpiry) {
		t.Errorf("expires at = %v, want %v", renewed.ExpiresAt, newExpiry)
	}

	remaining := storingCacheKeys(cache)
	if remaining["offering:"+testOfferingID.String()] {
		t.Error("the single-offering key survived a renewal")
	}
	if remaining["offerings:byuser:"+testUserID.String()] {
		t.Errorf("the farmer list key survived a renewal: %v", remaining)
	}
}

func TestCachedOfferingRenewKeepsEveryEntryWhenTheCallerIsRefused(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedOfferingUC(cache)

	warmOfferingCache(t, uc)

	_, err := uc.RenewOffering(principalCtxFor(testOtherID), testOfferingID, dto.RenewOfferingRequest{
		ExpiresAt: fixedTime.Add(15 * 24 * time.Hour),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RenewOffering() error = %v, want %v", err, domain.ErrForbidden)
	}

	if len(storingCacheKeys(cache)) != 2 {
		t.Errorf("a refused renewal dropped cache entries: %v", storingCacheKeys(cache))
	}
}
