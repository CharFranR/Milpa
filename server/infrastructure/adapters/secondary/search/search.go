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
}

func NewElasticSearchImpl(client *elasticsearch.Client) ElasticSearchImpl {
	return ElasticSearchImpl{client: client}
}

func (E *ElasticSearchImpl) Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error) {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": map[string]interface{}{
					"query":     term,   // Misspelled term
					"fuzziness": "AUTO", // Automatic fuzziness
				},
			},
		},
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {

		return nil, fmt.Errorf("Search: Error encoding the query: %w", err)
	}

	// Perform the search
	res, err := E.client.Search(
		E.client.Search.WithIndex("products"),
		E.client.Search.WithBody(&buf),
		E.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {

		return nil, fmt.Errorf("Search: Error in response  = %w", err)
	}

	defer res.Body.Close()

	// decocode the response

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

	// map the response

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
