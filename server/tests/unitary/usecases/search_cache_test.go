package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
)

func TestSearchCacheKeyDeterministic(t *testing.T) {
	t.Parallel()

	req := dto.SearchRequest{
		Term:       "maiz",
		Type:       "product",
		Department: "Managua",
		Page:       1,
		PageSize:   20,
	}

	key1 := usecases.SearchCacheKey(req)
	key2 := usecases.SearchCacheKey(req)

	if key1 != key2 {
		t.Errorf("expected same key for same request, got %q and %q", key1, key2)
	}
}

func TestSearchCacheKeyDifferent(t *testing.T) {
	t.Parallel()

	req1 := dto.SearchRequest{Term: "maiz", Page: 1}
	req2 := dto.SearchRequest{Term: "frijol", Page: 1}

	key1 := usecases.SearchCacheKey(req1)
	key2 := usecases.SearchCacheKey(req2)

	if key1 == key2 {
		t.Errorf("expected different keys for different requests, both got %q", key1)
	}
}

func TestCachedSearchUseCaseCacheMiss(t *testing.T) {
	t.Parallel()

	fuzzy := &fakeFuzzySearch{}
	c := newFakeCache()

	uc := usecases.NewCachedSearchUseCase(fuzzy, c)

	req := dto.SearchRequest{Term: "maiz", Page: 1, PageSize: 20}
	resp, err := uc.Search(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if !c.calledGet {
		t.Error("expected cache.Get to be called")
	}
	if !c.calledSet {
		t.Error("expected cache.Set to be called on miss")
	}
	if fuzzy.searchCount != 1 {
		t.Errorf("expected fuzzy search called once, got %d", fuzzy.searchCount)
	}
}

func TestCachedSearchUseCaseCacheHit(t *testing.T) {
	t.Parallel()

	fuzzy := &fakeFuzzySearch{}
	c := newFakeCache()

	cached := &dto.SearchResponse{
		Results:   []dto.FuzzySearchDto{{ID: "1", Name: "Maiz Dulce"}},
		TotalHits: 1,
		Page:      1,
		PageSize:  20,
	}
	c.getDest = cached

	uc := usecases.NewCachedSearchUseCase(fuzzy, c)

	req := dto.SearchRequest{Term: "maiz", Page: 1, PageSize: 20}
	resp, err := uc.Search(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Errorf("expected 1 total hit from cache, got %d", resp.TotalHits)
	}
	if fuzzy.searchCount != 0 {
		t.Errorf("expected fuzzy search NOT called on cache hit, got %d", fuzzy.searchCount)
	}
}

func TestCachedSearchUseCaseInvalidateAll(t *testing.T) {
	t.Parallel()

	c := newFakeCache()
	fuzzy := &fakeFuzzySearch{}
	uc := usecases.NewCachedSearchUseCase(fuzzy, c)

	err := uc.InvalidateAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.calledDeleteByPrefix {
		t.Error("expected DeleteByPrefix to be called")
	}
	if c.deletedPrefix != "search:" {
		t.Errorf("expected prefix 'search:', got %q", c.deletedPrefix)
	}
}

func TestCachedSearchUseCaseFuzzyError(t *testing.T) {
	t.Parallel()

	fuzzy := &fakeFuzzySearch{err: errors.New("es down")}
	c := newFakeCache()

	uc := usecases.NewCachedSearchUseCase(fuzzy, c)

	req := dto.SearchRequest{Term: "maiz", Page: 1}
	_, err := uc.Search(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from fuzzy, got nil")
	}
	if c.calledSet {
		t.Error("expected cache.Set NOT called on error")
	}
}

func TestOfferingInvalidatesSearchCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action func(uc *usecases.OfferingUseCaseImpl, ctx context.Context) error
	}{
		{
			name: "create offering invalidates",
			action: func(uc *usecases.OfferingUseCaseImpl, ctx context.Context) error {
				_, err := uc.CreateOffering(ctx, dto.CreateOfferingRequest{
					UserID: testUserID,
					Type:   1,
					Name:   "Test",
				})
				return err
			},
		},
		{
			name: "update offering invalidates",
			action: func(uc *usecases.OfferingUseCaseImpl, ctx context.Context) error {
				return uc.UpdateOffering(ctx, testOfferingID, dto.UpdateOfferingRequest{
					Name: strPtr("Updated"),
				})
			},
		},
		{
			name: "delete offering invalidates",
			action: func(uc *usecases.OfferingUseCaseImpl, ctx context.Context) error {
				return uc.DeleteOffering(ctx, testOfferingID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv := &fakeInvalidator{}
			uc := usecases.NewOfferingUseCase(
				newFakeOfferingRepo(),
				newFakeUserRepo(),
				newFakeTimer(),
				&fakeFuzzyRetrival{},
				inv,
			)

			err := tt.action(uc, principalCtx())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !inv.called {
				t.Error("expected InvalidateAll to be called")
			}
		})
	}
}

type fakeFuzzySearch struct {
	err         error
	searchCount int
}

func (f *fakeFuzzySearch) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	f.searchCount++
	if f.err != nil {
		return nil, f.err
	}
	return &dto.SearchResponse{
		Results:   []dto.FuzzySearchDto{},
		TotalHits: 0,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

type fakeCache struct {
	calledGet           bool
	calledSet           bool
	calledDelete        bool
	calledDeleteByPrefix bool
	deletedPrefix       string
	deletedKey          string
	getDest             any
	getFound            bool
}

func newFakeCache() *fakeCache {
	return &fakeCache{getFound: false}
}

func (c *fakeCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	c.calledGet = true
	if c.getDest != nil {
		switch d := dest.(type) {
		case **dto.SearchResponse:
			*d = c.getDest.(*dto.SearchResponse)
		}
		return true, nil
	}
	return false, nil
}

func (c *fakeCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.calledSet = true
	return nil
}

func (c *fakeCache) Delete(ctx context.Context, key string) error {
	c.calledDelete = true
	c.deletedKey = key
	return nil
}

func (c *fakeCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.calledDeleteByPrefix = true
	c.deletedPrefix = prefix
	return nil
}

func (c *fakeCache) Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error) {
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
