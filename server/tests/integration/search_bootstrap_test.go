package integration

import (
	"context"
	"encoding/json"
	"testing"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	esAdapter "milpa/infrastructure/adapters/secondary/search"

	"github.com/google/uuid"
)

const bootstrapIndex = "bootstrap-offerings"

// TestEnsureIndexCreatesIndexWithMapping proves the index the search path
// targets actually exists after boot, and that the mapping it was created with
// declares the geo_point the proximity sort resolves against.
func TestEnsureIndexCreatesIndexWithMapping(t *testing.T) {
	ctx := context.Background()

	deleteIndex(t, bootstrapIndex)

	if err := esAdapter.EnsureIndex(ctx, TestESClient, bootstrapIndex); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}
	defer deleteIndex(t, bootstrapIndex)

	mapping := indexMapping(t, bootstrapIndex)

	location, ok := mapping["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("mapping has no location property, got %v", mapping)
	}
	if location["type"] != "geo_point" {
		t.Errorf("location type = %v, want geo_point", location["type"])
	}

	// Every field the search filters and sorts on has to be an explicit type:
	// Elasticsearch otherwise infers it from the first document it sees and
	// silently mis-types every later one.
	for _, field := range []string{"id", "name", "description", "price", "type", "category_id", "user_id", "farmer_name", "farmer_verified", "department", "municipality", "latitude", "longitude", "location"} {
		if _, ok := mapping[field]; !ok {
			t.Errorf("mapping is missing the %q property", field)
		}
	}
}

// TestEnsureIndexIsIdempotent runs the bootstrap twice over an index that holds
// real documents and proves the second run neither errors nor loses data.
func TestEnsureIndexIsIdempotent(t *testing.T) {
	ctx := context.Background()

	deleteIndex(t, bootstrapIndex)
	defer deleteIndex(t, bootstrapIndex)

	if err := esAdapter.EnsureIndex(ctx, TestESClient, bootstrapIndex); err != nil {
		t.Fatalf("first EnsureIndex() error: %v", err)
	}

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, bootstrapIndex)

	seeded := []dto.IndexOfferingRequest{
		{
			ID: uuid.NewString(), Name: "Maiz Dulce Organico", Price: 15.0, Type: "product",
			Department: "Leon", Municipality: "Leon", Latitude: 12.4379, Longitude: -86.6161,
		},
		{
			ID: uuid.NewString(), Name: "Frijol Rojo Premium", Price: 25.0, Type: "product",
			Department: "Jinotega", Municipality: "Jinotega", Latitude: 13.0913, Longitude: -86.0014,
		},
	}
	for i := range seeded {
		if err := adapter.Index(ctx, &seeded[i]); err != nil {
			t.Fatalf("seed document %s: %v", seeded[i].Name, err)
		}
	}
	refreshIndex(t, bootstrapIndex)

	before := documentCount(t, bootstrapIndex)
	if before != int64(len(seeded)) {
		t.Fatalf("documents after seeding = %d, want %d", before, len(seeded))
	}

	// Second boot: the index is already there, so this has to be a no-op.
	if err := esAdapter.EnsureIndex(ctx, TestESClient, bootstrapIndex); err != nil {
		t.Fatalf("second EnsureIndex() error: %v", err)
	}

	after := documentCount(t, bootstrapIndex)
	if after != before {
		t.Errorf("documents after a second bootstrap = %d, want %d (no data loss)", after, before)
	}
}

// TestEnsureIndexRejectsEmptyIndexName guards the original defect: a blank
// index name is not a valid target, and silently accepting one is how every
// search call ended up pointed at nothing.
func TestEnsureIndexRejectsEmptyIndexName(t *testing.T) {
	err := esAdapter.EnsureIndex(context.Background(), TestESClient, "")
	if err == nil {
		t.Fatal("EnsureIndex(\"\") = nil, want an error for a blank index name")
	}
}

// TestSearchProximitySortResolvesGeoPoint is the end-to-end evidence that the
// proximity sort resolves: the index is bootstrapped from the embedded mapping,
// documents are indexed with coordinates, and a proximity-sorted search runs
// against the real container without a mapping error.
func TestSearchProximitySortResolvesGeoPoint(t *testing.T) {
	ctx := context.Background()

	deleteIndex(t, bootstrapIndex)
	defer deleteIndex(t, bootstrapIndex)

	if err := esAdapter.EnsureIndex(ctx, TestESClient, bootstrapIndex); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, bootstrapIndex)
	uc := usecases.NewSearchImpl(adapter)

	// Ordered by how far each producer is from Managua: Masaya is about 19 km,
	// Tipitapa about 45 km, Leon about 78 km.
	seeded := []dto.IndexOfferingRequest{
		{
			ID: uuid.NewString(), Name: "Cafe de Masaya", Price: 40.0, Type: "product",
			Department: "Masaya", Municipality: "Masateva", Latitude: 11.9747, Longitude: -86.0941,
		},
		{
			ID: uuid.NewString(), Name: "Queso de Tipitapa", Price: 30.0, Type: "product",
			Department: "Tipitapa", Municipality: "Tipitapa", Latitude: 12.5050, Longitude: -86.3620,
		},
		{
			ID: uuid.NewString(), Name: "Frijol de Leon", Price: 20.0, Type: "product",
			Department: "Leon", Municipality: "Leon", Latitude: 12.4379, Longitude: -86.8781,
		},
	}
	for i := range seeded {
		if err := adapter.Index(ctx, &seeded[i]); err != nil {
			t.Fatalf("seed document %s: %v", seeded[i].Name, err)
		}
	}
	refreshIndex(t, bootstrapIndex)

	// The indexed document has to carry the geo_point the sort measures.
	if location := documentLocation(t, bootstrapIndex, seeded[0].ID); location == nil {
		t.Fatalf("document %s has no geo_point, want one derived from its coordinates", seeded[0].Name)
	}

	resp, err := uc.Search(ctx, dto.SearchRequest{
		Term:      "",
		SortBy:    dto.SortProximity,
		Latitude:  12.1150,
		Longitude: -86.2362,
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("proximity-sorted search failed: %v", err)
	}
	if resp.TotalHits != int64(len(seeded)) {
		t.Errorf("total hits = %d, want %d", resp.TotalHits, len(seeded))
	}
	if len(resp.Results) != len(seeded) {
		t.Fatalf("results = %d, want %d", len(resp.Results), len(seeded))
	}

	wantOrder := []string{"Cafe de Masaya", "Queso de Tipitapa", "Frijol de Leon"}
	for i, want := range wantOrder {
		if resp.Results[i].Name != want {
			t.Errorf("result %d = %q, want %q (proximity order)", i, resp.Results[i].Name, want)
		}
	}
}

// TestSearchProximitySortWithoutCoordinatesStillRuns pins that a proximity sort
// with no origin does not become a mapping error: Elasticsearch excludes the
// documents that carry no geo_point and returns the rest.
func TestSearchProximitySortWithoutCoordinatesStillRuns(t *testing.T) {
	ctx := context.Background()

	deleteIndex(t, bootstrapIndex)
	defer deleteIndex(t, bootstrapIndex)

	if err := esAdapter.EnsureIndex(ctx, TestESClient, bootstrapIndex); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}

	adapter := esAdapter.NewElasticSearchImpl(TestESClient, bootstrapIndex)
	uc := usecases.NewSearchImpl(adapter)

	located := dto.IndexOfferingRequest{
		ID: uuid.NewString(), Name: "Frijol de Jinotega", Price: 25.0, Type: "product",
		Department: "Jinotega", Municipality: "Jinotega", Latitude: 13.0913, Longitude: -86.0014,
	}
	if err := adapter.Index(ctx, &located); err != nil {
		t.Fatalf("index located document: %v", err)
	}

	// No coordinates at all: the geo_point must be omitted rather than planted
	// at null island, and the document must not stop being searchable.
	unlocated := dto.IndexOfferingRequest{
		ID: uuid.NewString(), Name: "Servicio de Transporte", Price: 100.0, Type: "service",
		Department: "Managua", Municipality: "Managua",
	}
	if err := adapter.Index(ctx, &unlocated); err != nil {
		t.Fatalf("index unlocated document: %v", err)
	}
	refreshIndex(t, bootstrapIndex)

	if location := documentLocation(t, bootstrapIndex, unlocated.ID); location != nil {
		t.Errorf("document without coordinates got geo_point %v, want none", location)
	}

	resp, err := uc.Search(ctx, dto.SearchRequest{
		Term:     "",
		SortBy:   dto.SortProximity,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("proximity-sorted search failed: %v", err)
	}
	if resp.TotalHits == 0 {
		t.Fatal("total hits = 0, want the located document to remain searchable")
	}
	if resp.Results[0].Name != "Frijol de Jinotega" {
		t.Errorf("first result = %q, want the document that carries a geo_point", resp.Results[0].Name)
	}
}

func deleteIndex(t *testing.T, index string) {
	t.Helper()
	res, err := TestESClient.Indices.Delete([]string{index})
	if err != nil {
		t.Fatalf("delete index %s: %v", index, err)
	}
	defer res.Body.Close()
	// 404 is the normal case on the first boot.
	if res.IsError() && res.StatusCode != 404 {
		t.Fatalf("delete index %s: %s", index, res.Status())
	}
}

func refreshIndex(t *testing.T, index string) {
	t.Helper()
	res, err := TestESClient.Indices.Refresh(TestESClient.Indices.Refresh.WithIndex(index))
	if err != nil {
		t.Fatalf("refresh index %s: %v", index, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		t.Fatalf("refresh index %s: %s", index, res.Status())
	}
}

func indexMapping(t *testing.T, index string) map[string]interface{} {
	t.Helper()

	res, err := TestESClient.Indices.GetMapping(TestESClient.Indices.GetMapping.WithIndex(index))
	if err != nil {
		t.Fatalf("get mapping for %s: %v", index, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		t.Fatalf("get mapping for %s: %s", index, res.Status())
	}

	var payload map[string]struct {
		Mappings struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"mappings"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode mapping for %s: %v", index, err)
	}

	entry, ok := payload[index]
	if !ok {
		t.Fatalf("mapping response has no entry for %s", index)
	}

	return entry.Mappings.Properties
}

func documentCount(t *testing.T, index string) int64 {
	t.Helper()

	res, err := TestESClient.Count(TestESClient.Count.WithIndex(index))
	if err != nil {
		t.Fatalf("count documents in %s: %v", index, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		t.Fatalf("count documents in %s: %s", index, res.Status())
	}

	var payload struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode count for %s: %v", index, err)
	}

	return payload.Count
}

func documentLocation(t *testing.T, index, id string) map[string]interface{} {
	t.Helper()

	res, err := TestESClient.Get(index, id)
	if err != nil {
		t.Fatalf("get document %s: %v", id, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		t.Fatalf("get document %s: %s", id, res.Status())
	}

	var payload struct {
		Source struct {
			Location map[string]interface{} `json:"location"`
		} `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode document %s: %v", id, err)
	}

	return payload.Source.Location
}
