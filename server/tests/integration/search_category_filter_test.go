package integration

import (
	"context"
	"testing"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	esAdapter "milpa/infrastructure/adapters/secondary/search"

	"github.com/google/uuid"
)

const categoryIndex = "category-offerings"

var (
	fruitCategoryID  = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	citrusCategoryID = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
)

func seedCategoryIndex(t *testing.T, ctx context.Context) *usecases.SearchImpl {
	t.Helper()

	deleteIndex(t, categoryIndex)
	t.Cleanup(func() { deleteIndex(t, categoryIndex) })

	if err := esAdapter.EnsureIndex(ctx, TestESClient, categoryIndex); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, categoryIndex)

	seeded := []dto.IndexOfferingRequest{
		{
			ID: uuid.NewString(), Name: "Mango Julie", Price: 15.0, Type: "product",
			CategoryID: fruitCategoryID.String(),
			Department: "Masaya", Municipality: "Masateva", Latitude: 11.9747, Longitude: -86.0941,
		},
		{
			ID: uuid.NewString(), Name: "Naranja Valencia", Price: 22.0, Type: "product",
			CategoryID: citrusCategoryID.String(),
			Department: "Jinotega", Municipality: "Jinotega", Latitude: 13.0913, Longitude: -86.0014,
		},
		{
			ID: uuid.NewString(), Name: "Frijol Rojo", Price: 30.0, Type: "product",
			Department: "Leon", Municipality: "Leon", Latitude: 12.4379, Longitude: -86.8781,
		},
	}

	for i := range seeded {
		if err := adapter.Index(ctx, &seeded[i]); err != nil {
			t.Fatalf("seed document %s: %v", seeded[i].Name, err)
		}
	}
	refreshIndex(t, categoryIndex)

	return usecases.NewSearchImpl(adapter)
}

func searchNames(resp *dto.SearchResponse) map[string]bool {
	names := make(map[string]bool, len(resp.Results))
	for _, result := range resp.Results {
		names[result.Name] = true
	}
	return names
}

func TestSearchFiltersByCatalogueCategory(t *testing.T) {
	ctx := context.Background()

	uc := seedCategoryIndex(t, ctx)

	matched, err := uc.Search(ctx, dto.SearchRequest{CategoryID: fruitCategoryID.String(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search by category: %v", err)
	}
	if matched.TotalHits != 1 {
		t.Fatalf("hits for %s = %d, want 1", fruitCategoryID, matched.TotalHits)
	}
	if names := searchNames(matched); !names["Mango Julie"] {
		t.Errorf("the category filter returned %v, want the fruit product", names)
	}

	other, err := uc.Search(ctx, dto.SearchRequest{CategoryID: citrusCategoryID.String(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search by the other category: %v", err)
	}
	if other.TotalHits != 1 {
		t.Fatalf("hits for %s = %d, want 1", citrusCategoryID, other.TotalHits)
	}
	if names := searchNames(other); !names["Naranja Valencia"] {
		t.Errorf("the category filter returned %v, want the citrus product", names)
	}

	unknown, err := uc.Search(ctx, dto.SearchRequest{CategoryID: uuid.NewString(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search by an uncategorised id: %v", err)
	}
	if unknown.TotalHits != 0 {
		t.Errorf("hits for an unknown category = %d, want 0", unknown.TotalHits)
	}
	if len(unknown.Results) != 0 {
		t.Errorf("results for an unknown category = %v, want none", unknown.Results)
	}
}

func TestSearchKeepsTheLegacyTypeParameter(t *testing.T) {
	ctx := context.Background()

	uc := seedCategoryIndex(t, ctx)

	resp, err := uc.Search(ctx, dto.SearchRequest{Type: "product", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search by the legacy type parameter: %v", err)
	}
	if resp.TotalHits != 3 {
		t.Errorf("hits for type=product = %d, want the 3 indexed products", resp.TotalHits)
	}

	combined, err := uc.Search(ctx, dto.SearchRequest{Type: "product", CategoryID: citrusCategoryID.String(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search by type and category: %v", err)
	}
	if combined.TotalHits != 1 {
		t.Fatalf("hits for type=product and the citrus category = %d, want 1", combined.TotalHits)
	}
	if names := searchNames(combined); !names["Naranja Valencia"] {
		t.Errorf("the combined filter returned %v, want the citrus product", names)
	}
}

func TestSearchCategoryFilterComposesWithTheTextSearch(t *testing.T) {
	ctx := context.Background()

	uc := seedCategoryIndex(t, ctx)

	resp, err := uc.Search(ctx, dto.SearchRequest{Term: "Naranja", CategoryID: citrusCategoryID.String(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("text search with a category filter: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Fatalf("hits = %d, want 1", resp.TotalHits)
	}

	wrongCategory, err := uc.Search(ctx, dto.SearchRequest{Term: "Naranja", CategoryID: fruitCategoryID.String(), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("text search with the wrong category: %v", err)
	}
	if wrongCategory.TotalHits != 0 {
		t.Errorf("hits = %d, want the text match excluded by the category", wrongCategory.TotalHits)
	}
}

func TestCategoryMappingIsAKeyword(t *testing.T) {
	ctx := context.Background()

	deleteIndex(t, categoryIndex)
	t.Cleanup(func() { deleteIndex(t, categoryIndex) })

	if err := esAdapter.EnsureIndex(ctx, TestESClient, categoryIndex); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}

	mapping := indexMapping(t, categoryIndex)

	category, ok := mapping["category_id"].(map[string]interface{})
	if !ok {
		t.Fatalf("mapping has no category_id property, got %v", mapping)
	}
	if category["type"] != "keyword" {
		t.Errorf("category_id type = %v, want keyword", category["type"])
	}
}
