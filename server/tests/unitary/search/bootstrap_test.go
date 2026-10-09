package search_test

import (
	"encoding/json"
	"testing"

	"milpa/aplication/dto"
	esAdapter "milpa/infrastructure/adapters/secondary/search"
)

// TestIndexMappingDeclaresGeoPoint checks the embedded mapping without needing
// a running Elasticsearch: the geo_point has to be declared, and every field the
// search filters, sorts or boosts on has to be typed explicitly so Elasticsearch
// cannot infer it from whichever document happens to land first.
func TestIndexMappingDeclaresGeoPoint(t *testing.T) {
	body, err := esAdapter.IndexMapping()
	if err != nil {
		t.Fatalf("IndexMapping() error: %v", err)
	}

	var definition struct {
		Settings map[string]interface{} `json:"settings"`
		Mappings struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(body, &definition); err != nil {
		t.Fatalf("decode embedded mapping: %v", err)
	}

	location, ok := definition.Mappings.Properties["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("mapping has no location property, got %v", definition.Mappings.Properties)
	}
	if location["type"] != "geo_point" {
		t.Errorf("location type = %v, want geo_point", location["type"])
	}

	wantTypes := map[string]string{
		"id":              "keyword",
		"name":            "text",
		"description":     "text",
		"price":           "float",
		"type":            "keyword",
		"user_id":         "keyword",
		"farmer_name":     "text",
		"farmer_verified": "boolean",
		"department":      "keyword",
		"municipality":    "keyword",
		"latitude":        "float",
		"longitude":       "float",
		"location":        "geo_point",
	}
	for field, wantType := range wantTypes {
		property, ok := definition.Mappings.Properties[field].(map[string]interface{})
		if !ok {
			t.Errorf("mapping is missing the %q property", field)
			continue
		}
		if property["type"] != wantType {
			t.Errorf("%s type = %v, want %s", field, property["type"], wantType)
		}
	}

	if definition.Settings["number_of_shards"] != float64(1) {
		t.Errorf("number_of_shards = %v, want 1 for a single-node deployment", definition.Settings["number_of_shards"])
	}
}

// TestNewGeoPoint covers the rule that decides whether a document carries a
// geo_point at all.
func TestNewGeoPoint(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		wantNil   bool
	}{
		{name: "real coordinates", latitude: 12.4379, longitude: -86.8781},
		{name: "no coordinates", latitude: 0, longitude: 0, wantNil: true},
		{name: "only latitude", latitude: 12.4379, longitude: 0},
		{name: "only longitude", latitude: 0, longitude: -86.8781},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dto.NewGeoPoint(tt.latitude, tt.longitude)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("NewGeoPoint(%v, %v) = %+v, want nil", tt.latitude, tt.longitude, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("NewGeoPoint(%v, %v) = nil, want a geo_point", tt.latitude, tt.longitude)
			}
			if got.Lat != tt.latitude || got.Lon != tt.longitude {
				t.Errorf("geo_point = %+v, want lat %v lon %v", got, tt.latitude, tt.longitude)
			}
		})
	}
}
