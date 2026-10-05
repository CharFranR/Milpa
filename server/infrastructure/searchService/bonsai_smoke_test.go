package elasticSsearch

import (
	"context"
	"os"
	"testing"
	"time"

	"milpa/aplication/dto"
	"milpa/infrastructure/adapters/secondary/search"
)

func TestBonsaiSmoke(t *testing.T) {
	endpoint := os.Getenv("TEST_ES_ENDPOINT")
	if endpoint == "" {
		t.Skip("TEST_ES_ENDPOINT not set")
	}
	client, err := CreateESClient(dto.ESClient{
		Endpoint1: endpoint,
		Username:  os.Getenv("TEST_ES_USER"),
		Password:  os.Getenv("TEST_ES_PASS"),
	})
	if err != nil {
		t.Fatalf("CreateESClient: %v", err)
	}
	index := "milpa-smoke-" + time.Now().Format("20060102150405")
	if err := search.EnsureIndex(context.Background(), client, index); err != nil {
		t.Fatalf("EnsureIndex: %v", err)
	}
	created, err := client.Indices.Exists([]string{index})
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	defer created.Body.Close()
	if created.StatusCode != 200 {
		t.Fatalf("index not found after EnsureIndex: %d", created.StatusCode)
	}
	deleted, err := client.Indices.Delete([]string{index})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	defer deleted.Body.Close()
}
