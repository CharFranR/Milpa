package usecases_test

import (
	"context"
	"testing"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
)

type capturingFuzzySearch struct {
	query *dto.SearchQuery
}

func (f *capturingFuzzySearch) Search(ctx context.Context, query *dto.SearchQuery) (*dto.SearchResponse, error) {
	f.query = query
	return &dto.SearchResponse{}, nil
}

func TestSearchQueryCarriesTheCatalogueCategory(t *testing.T) {
	t.Parallel()

	fuzzy := &capturingFuzzySearch{}
	uc := usecases.NewSearchImpl(fuzzy)

	if _, err := uc.Search(context.Background(), dto.SearchRequest{
		CategoryID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Type:       "product",
	}); err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if fuzzy.query == nil {
		t.Fatal("the adapter was never called")
	}
	if fuzzy.query.Filters.CategoryID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Errorf("filter category = %q, want the requested category", fuzzy.query.Filters.CategoryID)
	}
	if fuzzy.query.Filters.Type != "product" {
		t.Errorf("filter type = %q, want the legacy type to survive alongside it", fuzzy.query.Filters.Type)
	}
}

func TestSearchQueryLeavesTheCategoryEmptyWhenUnset(t *testing.T) {
	t.Parallel()

	fuzzy := &capturingFuzzySearch{}
	uc := usecases.NewSearchImpl(fuzzy)

	if _, err := uc.Search(context.Background(), dto.SearchRequest{Term: "maiz"}); err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if fuzzy.query.Filters.CategoryID != "" {
		t.Errorf("filter category = %q, want it unset so no category filter is emitted", fuzzy.query.Filters.CategoryID)
	}
}
