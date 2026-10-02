package search

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

const IndexOfferingsMappingFile = "offerings.json"

const geoPointField = "location"

//go:embed mappings
var Mappings embed.FS

var ErrEmptyIndexName = errors.New("elasticsearch index name is empty")

// IndexMapping returns the raw index definition for the offerings index.
func IndexMapping() ([]byte, error) {
	body, err := Mappings.ReadFile("mappings/" + IndexOfferingsMappingFile)
	if err != nil {
		return nil, fmt.Errorf("EnsureIndex: read embedded mapping: %w", err)
	}
	return body, nil
}

func EnsureIndex(ctx context.Context, client *elasticsearch.Client, index string) error {
	if index == "" {
		return ErrEmptyIndexName
	}
	if client == nil {
		return errors.New("EnsureIndex: elasticsearch client is nil")
	}

	head, err := client.Indices.Exists([]string{index}, client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("EnsureIndex: check index %q: %w", index, err)
	}
	defer head.Body.Close()

	switch {
	case head.StatusCode == http.StatusOK:
		return nil
	case head.StatusCode != http.StatusNotFound:
		return fmt.Errorf("EnsureIndex: check index %q: %s", index, head.Status())
	}

	mapping, err := IndexMapping()
	if err != nil {
		return err
	}

	res, err := client.Indices.Create(index,
		client.Indices.Create.WithContext(ctx),
		client.Indices.Create.WithBody(strings.NewReader(string(mapping))),
	)
	if err != nil {
		return fmt.Errorf("EnsureIndex: create index %q: %w", index, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode == http.StatusBadRequest && bytesContainAlreadyExists(body) {
			return nil
		}
		return fmt.Errorf("EnsureIndex: create index %q: %s: %s", index, res.Status(), body)
	}

	return nil
}

func bytesContainAlreadyExists(body []byte) bool {
	var payload struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	return payload.Error.Type == "resource_already_exists_exception"
}
