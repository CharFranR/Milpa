package usecases_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

// storingCache is a real in-memory cache: it stores by key and reads it back
// through JSON, the way the Redis adapter does. The fake the search tests use is
// deliberately stateless, which cannot demonstrate anything about key scoping.
type storingCache struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newStoringCache() *storingCache {
	return &storingCache{items: map[string][]byte{}}
}

func (c *storingCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	raw, ok := c.items[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *storingCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = raw
	return nil
}

func (c *storingCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

func (c *storingCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
		}
	}
	return nil
}

func (c *storingCache) Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error) {
	found, err := c.Get(ctx, key, dest)
	if err != nil {
		return false, err
	}
	if found {
		return true, nil
	}
	if err := loader(); err != nil {
		return false, err
	}
	_ = c.Set(ctx, key, dest, ttl)
	return false, nil
}

var _ port.Cache = (*storingCache)(nil)

func cachedUserUC(c *storingCache) *usecases.CachedUserUseCase {
	userRepo := newFakeUserRepo()
	fixture := piiFixture()
	userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		return fixture, nil
	}
	return usecases.NewCachedUserUseCase(
		usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer()),
		c,
	)
}

func cachedCompanyUC(c *storingCache) *usecases.CachedCompanyUseCase {
	companyRepo := newFakeCompanyRepo()
	fixture := mustCompany()
	fixture.Email = "info@milpa.com.ni"
	fixture.PhoneNumber = "555-9999"
	fixture.Address = domain.Address{AddressLine: "Casa 12"}
	companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
		return fixture, nil
	}
	companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
		return []domain.Company{*fixture}, nil
	}
	return usecases.NewCachedCompanyUseCase(
		usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer()),
		c,
	)
}

// TestCachedUserGetByIDIsolatesViewers is the cache-shaped version of the same
// boundary. A cache keyed only by the user id would hand the owner's contact
// card to the next anonymous visitor for as long as the entry lived, which is
// the whole leak the DTO split was meant to close.
func TestCachedUserGetByIDIsolatesViewers(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedUserUC(cache)

	// The owner reads first and populates the cache with the private view.
	ownerView, err := uc.GetByID(principalCtx(), testUserID)
	if err != nil {
		t.Fatalf("owner read: %v", err)
	}
	if _, ok := ownerView.(*dto.PrivateUserDTO); !ok {
		t.Fatalf("owner read = %T, want the private view", ownerView)
	}

	// An anonymous visitor must not receive the owner's cached contact card.
	anonView, err := uc.GetByID(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("anonymous read: %v", err)
	}
	if _, leaked := anonView.(*dto.PrivateUserDTO); leaked {
		t.Fatal("the cache served the owner's private view to an anonymous caller")
	}

	anonEncoded, _ := json.Marshal(anonView)
	if strings.Contains(string(anonEncoded), "user@milpa.com.ni") {
		t.Errorf("anonymous read exposed the email: %s", anonEncoded)
	}

	// A third authenticated user is not entitled either.
	thirdView, err := uc.GetByID(principalCtxFor(testOtherID), testUserID)
	if err != nil {
		t.Fatalf("third party read: %v", err)
	}
	if _, leaked := thirdView.(*dto.PrivateUserDTO); leaked {
		t.Fatal("the cache served the owner's private view to a third party")
	}

	// And the owner still gets their own contact card on a warm cache.
	warmOwner, err := uc.GetByID(principalCtx(), testUserID)
	if err != nil {
		t.Fatalf("owner warm read: %v", err)
	}
	if _, ok := warmOwner.(*dto.PrivateUserDTO); !ok {
		t.Errorf("owner warm read = %T, want the private view", warmOwner)
	}
}

func TestCachedCompanyGetByIDIsolatesViewers(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedCompanyUC(cache)

	ownerView, err := uc.GetByID(principalCtx(), testCompanyID)
	if err != nil {
		t.Fatalf("owner read: %v", err)
	}
	if _, ok := ownerView.(*dto.PrivateCompanyDTO); !ok {
		t.Fatalf("owner read = %T, want the private view", ownerView)
	}

	anonView, err := uc.GetByID(context.Background(), testCompanyID)
	if err != nil {
		t.Fatalf("anonymous read: %v", err)
	}
	if _, leaked := anonView.(*dto.PrivateCompanyDTO); leaked {
		t.Fatal("the cache served the company's private view to an anonymous caller")
	}

	anonEncoded, _ := json.Marshal(anonView)
	for _, secret := range []string{"info@milpa.com.ni", "555-9999", "Casa 12"} {
		if strings.Contains(string(anonEncoded), secret) {
			t.Errorf("anonymous read exposed %q: %s", secret, anonEncoded)
		}
	}

	thirdView, err := uc.GetByID(principalCtxFor(testOtherID), testCompanyID)
	if err != nil {
		t.Fatalf("third party read: %v", err)
	}
	if _, leaked := thirdView.(*dto.PrivateCompanyDTO); leaked {
		t.Fatal("the cache served the company's private view to a third party")
	}

	warmOwner, err := uc.GetByID(principalCtx(), testCompanyID)
	if err != nil {
		t.Fatalf("owner warm read: %v", err)
	}
	if _, ok := warmOwner.(*dto.PrivateCompanyDTO); !ok {
		t.Errorf("owner warm read = %T, want the private view", warmOwner)
	}
}

func TestCachedCompanyGetByOwnerIsolatesViewers(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedCompanyUC(cache)

	ownerViews, err := uc.GetByOwner(principalCtx(), testUserID)
	if err != nil {
		t.Fatalf("owner read: %v", err)
	}
	if len(ownerViews) != 1 {
		t.Fatalf("owner read returned %d views, want 1", len(ownerViews))
	}
	if _, ok := ownerViews[0].(*dto.PrivateCompanyDTO); !ok {
		t.Fatalf("owner read = %T, want the private view", ownerViews[0])
	}

	anonViews, err := uc.GetByOwner(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("anonymous read: %v", err)
	}
	encoded, _ := json.Marshal(anonViews)
	if strings.Contains(string(encoded), "info@milpa.com.ni") {
		t.Errorf("anonymous listing exposed the company email: %s", encoded)
	}
	for i, view := range anonViews {
		if _, leaked := view.(*dto.PrivateCompanyDTO); leaked {
			t.Errorf("listing %d is the private view, want public for an anonymous caller", i)
		}
	}
}

// TestCachedUserUpdateInvalidatesEveryViewer checks the other direction: after
// a profile change no viewer may be served a stale entry, whichever slot it
// happened to land in.
func TestCachedUserUpdateInvalidatesEveryViewer(t *testing.T) {
	t.Parallel()

	cache := newStoringCache()
	uc := cachedUserUC(cache)

	if _, err := uc.GetByID(principalCtx(), testUserID); err != nil {
		t.Fatalf("owner read: %v", err)
	}
	if _, err := uc.GetByID(context.Background(), testUserID); err != nil {
		t.Fatalf("anonymous read: %v", err)
	}
	if _, err := uc.GetByID(principalCtxFor(testOtherID), testUserID); err != nil {
		t.Fatalf("third party read: %v", err)
	}

	if err := uc.UpdateProfile(principalCtx(), testUserID, dto.UpdateUserRequest{FirstName: strPtr("Carlos")}); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	cache.mu.Lock()
	remaining := len(cache.items)
	cache.mu.Unlock()
	if remaining != 0 {
		t.Errorf("cache entries after an update = %d, want 0", remaining)
	}
}
