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

// IndexOfferingsMappingFile is the name of the embedded index definition.
const IndexOfferingsMappingFile = "offerings.json"

// geoPointField is the single name the three places that must agree on the
// geospatial field use: the index mapping, the indexed document, and the
// proximity sort. They drifted apart once already, which is why the sort
// could not resolve a geo_point that no mapping ever declared.
const geoPointField = "location"

// Mappings carries the index definitions compiled into the binary.
//
// Embedding them is what makes booting independent of the working directory,
// the same reason the SQL migrations are embedded next to the migration runner:
// a path relative to the process would only resolve when the binary happened to
// be started from the module root.
//
//go:embed mappings/*.json
var Mappings embed.FS

// ErrEmptyIndexName rejects a blank index name instead of letting every
// search call target the whole cluster. An index name of "" is not a valid
// Elasticsearch target, and a caller that passes one has a wiring bug.
var ErrEmptyIndexName = errors.New("elasticsearch index name is empty")

// IndexMapping returns the raw index definition for the offerings index.
func IndexMapping() ([]byte, error) {
	body, err := Mappings.ReadFile("mappings/" + IndexOfferingsMappingFile)
	if err != nil {
		return nil, fmt.Errorf("EnsureIndex: read embedded mapping: %w", err)
	}
	return body, nil
}

// EnsureIndex creates index with its explicit mapping when it does not exist.
//
// It is safe to run on every boot: an index that already exists is left exactly
// as it is, so the documents already indexed survive. Nothing is deleted and
// nothing is reindexed — the only statement that ever mutates Elasticsearch is
// the create, and only when the index is absent.
//
// A concurrent creator is not a failure: Elasticsearch answers a racing create
// with resource_already_exists_exception, and from this caller's point of view
// the postcondition (the index exists) already holds.
func EnsureIndex(ctx context.Context, client *elasticsearch.Client, index string) error {
	if index == "" {
		return ErrEmptyIndexName
	}
	if client == nil {
		return errors.New("EnsureIndex: elasticsearch client is nil")
	}

	// HEAD is the cheapest existence probe: 200 means the index is already
	// there and must be left alone, 404 means it has to be created.
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

// bytesContainAlreadyExists distinguishes "someone else won the race" from a
// genuinely malformed mapping. Only the former may be swallowed.
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
