package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"milpa/aplication/dto"
	cacheAdapter "milpa/infrastructure/adapters/secondary/cache"
	esAdapter "milpa/infrastructure/adapters/secondary/search"
	"milpa/aplication/use-cases"

	"github.com/google/uuid"
)

const testIndex = "test-offerings"

var testSetupDone bool

func setupSearchIndex(t *testing.T) {
	t.Helper()
	if testSetupDone {
		return
	}

	ctx := context.Background()

	// Delete index if exists
	TestESClient.Indices.Delete([]string{testIndex})

	// Create index with mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":               map[string]interface{}{"type": "keyword"},
				"name":             map[string]interface{}{"type": "text", "fields": map[string]interface{}{"keyword": map[string]interface{}{"type": "keyword"}}},
				"description":      map[string]interface{}{"type": "text"},
				"price":            map[string]interface{}{"type": "float"},
				"type":             map[string]interface{}{"type": "keyword"},
				"image_url":        map[string]interface{}{"type": "keyword", "index": false},
				"user_id":          map[string]interface{}{"type": "keyword"},
				"farmer_name":      map[string]interface{}{"type": "text"},
				"farmer_verified":  map[string]interface{}{"type": "boolean"},
				"department":       map[string]interface{}{"type": "keyword"},
				"municipality":     map[string]interface{}{"type": "keyword"},
				"location":         map[string]interface{}{"type": "geo_point"},
				"created_at":       map[string]interface{}{"type": "date"},
			},
		},
	}

	body, _ := json.Marshal(mapping)
	res, err := TestESClient.Indices.Create(
		testIndex,
		TestESClient.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		t.Fatalf("failed to create index: %v", err)
	}
	if res.IsError() {
		t.Fatalf("failed to create index: %s", res.String())
	}

	// Seed test data
	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)

	seedData := []dto.IndexOfferingRequest{
		{
			ID: uuid.New().String(), Name: "Maiz Dulce Organico", Description: "Maiz fresco de Leon",
			Price: 15.0, Type: "product", UserID: uuid.New().String(), FarmerName: "Carlos Verified",
			FarmerVerified: true, Department: "Leon", Municipality: "Leon",
			Latitude: 12.4379, Longitude: -86.6161,
		},
		{
			ID: uuid.New().String(), Name: "Frijol Rojo Premium", Description: "Frijol rojo de Jinotega",
			Price: 25.0, Type: "product", UserID: uuid.New().String(), FarmerName: "Maria Verified",
			FarmerVerified: true, Department: "Jinotega", Municipality: "Jinotega",
			Latitude: 13.0913, Longitude: -86.0014,
		},
		{
			ID: uuid.New().String(), Name: "Cafe Maragogype", Description: "Cafe de alta calidad Matagalpa",
			Price: 40.0, Type: "product", UserID: uuid.New().String(), FarmerName: "Pedro Unverified",
			FarmerVerified: false, Department: "Matagalpa", Municipality: "Matagalpa",
			Latitude: 12.9248, Longitude: -85.9171,
		},
		{
			ID: uuid.New().String(), Name: "Servicio de Transporte", Description: "Transporte de cosechas",
			Price: 100.0, Type: "service", UserID: uuid.New().String(), FarmerName: "Ana Unverified",
			FarmerVerified: false, Department: "Managua", Municipality: "Managua",
			Latitude: 12.1150, Longitude: -86.2362,
		},
		{
			ID: uuid.New().String(), Name: "Arroz Verde", Description: "Arroz organico Jinotega",
			Price: 18.0, Type: "product", UserID: uuid.New().String(), FarmerName: "Luis Verified",
			FarmerVerified: true, Department: "Jinotega", Municipality: "Wiwili",
			Latitude: 13.3833, Longitude: -85.8167,
		},
	}

	for _, d := range seedData {
		if err := adapter.Index(ctx, &d); err != nil {
			t.Fatalf("failed to seed document %s: %v", d.Name, err)
		}
	}

	// Wait for ES to refresh
	time.Sleep(1 * time.Second)
	testSetupDone = true
}

func TestSearchPagination(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	// Page 1, size 2
	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "",
		Page:     1,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 5 {
		t.Errorf("expected 5 total hits, got %d", resp.TotalHits)
	}
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(resp.Results))
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", resp.TotalPages)
	}

	// Page 2
	resp2, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "",
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp2.Results) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(resp2.Results))
	}

	// Page 3 (only 1 remaining)
	resp3, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "",
		Page:     3,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp3.Results) != 1 {
		t.Errorf("expected 1 result on page 3, got %d", len(resp3.Results))
	}
}

func TestSearchFilterDepartment(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:       "",
		Department: "Jinotega",
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 2 {
		t.Errorf("expected 2 results in Jinotega, got %d", resp.TotalHits)
	}
	for _, r := range resp.Results {
		if r.Department != "Jinotega" {
			t.Errorf("expected department Jinotega, got %s", r.Department)
		}
	}
}

func TestSearchFilterMunicipality(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:         "",
		Municipality: "Leon",
		Page:         1,
		PageSize:     20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Errorf("expected 1 result in Leon municipality, got %d", resp.TotalHits)
	}
	if resp.Results[0].Municipality != "Leon" {
		t.Errorf("expected municipality Leon, got %s", resp.Results[0].Municipality)
	}
}

func TestSearchFilterPriceRange(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	minPrice := 20.0
	maxPrice := 50.0
	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "",
		PriceMin: &minPrice,
		PriceMax: &maxPrice,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 2 {
		t.Errorf("expected 2 results in price range [20,50], got %d", resp.TotalHits)
	}
	for _, r := range resp.Results {
		if r.Price < 20.0 || r.Price > 50.0 {
			t.Errorf("expected price in [20,50], got %f", r.Price)
		}
	}
}

func TestSearchFilterType(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "",
		Type:     "service",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Errorf("expected 1 service result, got %d", resp.TotalHits)
	}
	if resp.Results[0].Type != "service" {
		t.Errorf("expected type service, got %s", resp.Results[0].Type)
	}
}

func TestSearchVerifiedBoost(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	// Search for "Organico" — should return Carlos Verified (boosted) before others
	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "Organico",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits < 2 {
		t.Fatalf("expected at least 2 results, got %d", resp.TotalHits)
	}

	// First result should be the verified farmer
	if !resp.Results[0].FarmerVerified {
		t.Errorf("expected first result to be verified farmer, got unverified: %s", resp.Results[0].FarmerName)
	}
}

func TestSearchTextMatch(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "frijol",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Errorf("expected 1 result for 'frijol', got %d", resp.TotalHits)
	}
	if resp.Results[0].Name != "Frijol Rojo Premium" {
		t.Errorf("expected 'Frijol Rojo Premium', got '%s'", resp.Results[0].Name)
	}
}

func TestSearchNoResults(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	uc := usecases.NewSearchImpl(adapter)

	resp, err := uc.Search(context.Background(), dto.SearchRequest{
		Term:     "xyznoexist",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalHits != 0 {
		t.Errorf("expected 0 results, got %d", resp.TotalHits)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results slice, got %d items", len(resp.Results))
	}
}

func TestSearchCacheIntegration(t *testing.T) {
	setupSearchIndex(t)

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, testIndex)
	searchUC := usecases.NewSearchImpl(adapter)

	// Create real Redis cache
	cacheClient := cacheAdapter.NewCacheImpl(TestRedisAddr, "", 0)

	cachedUC := usecases.NewCachedSearchUseCase(searchUC, cacheClient)

	req := dto.SearchRequest{Term: "Maiz", Page: 1, PageSize: 10}

	// First call — cache miss
	start := time.Now()
	resp1, err := cachedUC.Search(context.Background(), req)
	missDuration := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.TotalHits == 0 {
		t.Fatal("expected results for 'Maiz'")
	}

	// Second call — cache hit (should be faster)
	start = time.Now()
	resp2, err := cachedUC.Search(context.Background(), req)
	hitDuration := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.TotalHits != resp1.TotalHits {
		t.Errorf("expected same total hits, got %d vs %d", resp1.TotalHits, resp2.TotalHits)
	}

	t.Logf("cache miss: %v, cache hit: %v", missDuration, hitDuration)

	// Invalidate and verify miss again
	if err := cachedUC.InvalidateAll(context.Background()); err != nil {
		t.Fatalf("failed to invalidate: %v", err)
	}

	resp3, err := cachedUC.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error after invalidation: %v", err)
	}
	if resp3.TotalHits != resp1.TotalHits {
		t.Errorf("expected same total hits after invalidation, got %d vs %d", resp1.TotalHits, resp3.TotalHits)
	}
}

func TestSearchHTTPEndpoint(t *testing.T) {
	setupSearchIndex(t)

	// Test the HTTP endpoint directly using the real ES container
	url := fmt.Sprintf("%s/_search", TestESEndpoint)

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 2,
		"from": 0,
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("GET", url, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if esResp.Hits.Total.Value != 5 {
		t.Errorf("expected 5 total hits, got %d", esResp.Hits.Total.Value)
	}
	if len(esResp.Hits.Hits) != 2 {
		t.Errorf("expected 2 hits, got %d", len(esResp.Hits.Hits))
	}
}
