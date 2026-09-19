package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"milpa/aplication/dto"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticSearchImpl struct {
	client *elasticsearch.Client
	index  string
}

func NewElasticSearchImpl(client *elasticsearch.Client, index string) *ElasticSearchImpl {
	return &ElasticSearchImpl{client: client, index: index}
}

func (E *ElasticSearchImpl) Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error) {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": map[string]interface{}{
					"query":     term,
					"fuzziness": "AUTO",
				},
			},
		},
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("Search: Error encoding the query: %w", err)
	}

	res, err := E.client.Search(
		E.client.Search.WithIndex(E.index),
		E.client.Search.WithBody(&buf),
		E.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, fmt.Errorf("Search: Error in response = %w", err)
	}

	defer res.Body.Close()

	var esResp struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("Search: Error decoding response: %w", err)
	}

	result := make([]dto.FuzzySearchDto, 0, len(esResp.Hits.Hits))

	for _, h := range esResp.Hits.Hits {
		var dtoItem dto.FuzzySearchDto
		if err := json.Unmarshal(h.Source, &dtoItem); err != nil {
			return nil, fmt.Errorf("Search: error decoding source: %w", err)
		}
		result = append(result, dtoItem)
	}

	return result, nil
}

func (E *ElasticSearchImpl) Index(ctx context.Context, p *dto.IndexOfferingRequest) error {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Index: Error in json marshal: %w", err)
	}

	_, err = E.client.Index(
		E.index,
		bytes.NewReader(jsonData),
		E.client.Index.WithDocumentID(p.ID),
		E.client.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("Index: Error indexing document: %w", err)
	}

	return nil
}

func (E *ElasticSearchImpl) Update(ctx context.Context, id string, p *dto.IndexOfferingRequest) error {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Update: Error in json marshal: %w", err)
	}

	doc := map[string]interface{}{
		"doc":         json.RawMessage(jsonData),
		"doc_as_upsert": true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("Update: Error encoding update body: %w", err)
	}

	_, err = E.client.Update(
		E.index,
		id,
		&buf,
		E.client.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("Update: Error updating document %s: %w", id, err)
	}

	return nil
}

func (E *ElasticSearchImpl) Delete(ctx context.Context, id string) error {
	_, err := E.client.Delete(
		E.index,
		id,
		E.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("Delete: Error deleting document %s: %w", id, err)
	}

	return nil
}
