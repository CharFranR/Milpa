package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
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

// Search translates a SearchQuery into ES DSL. Pure translation — no business logic.
func (E *ElasticSearchImpl) Search(ctx context.Context, query *dto.SearchQuery) (*dto.SearchResponse, error) {
	esQuery := E.buildQuery(query)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, fmt.Errorf("Search: error encoding query: %w", err)
	}

	from := (query.Pagination.Page - 1) * query.Pagination.PageSize

	res, err := E.client.Search(
		E.client.Search.WithIndex(E.index),
		E.client.Search.WithBody(&buf),
		E.client.Search.WithTrackTotalHits(true),
		E.client.Search.WithFrom(from),
		E.client.Search.WithSize(query.Pagination.PageSize),
	)
	if err != nil {
		return nil, fmt.Errorf("Search: error executing query: %w", err)
	}
	defer res.Body.Close()

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string          `json:"_id"`
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("Search: error decoding response: %w", err)
	}

	results := make([]dto.FuzzySearchDto, 0, len(esResp.Hits.Hits))
	for _, h := range esResp.Hits.Hits {
		var item dto.FuzzySearchDto
		if err := json.Unmarshal(h.Source, &item); err != nil {
			return nil, fmt.Errorf("Search: error decoding source: %w", err)
		}
		results = append(results, item)
	}

	totalPages := 0
	if query.Pagination.PageSize > 0 {
		totalPages = int(math.Ceil(float64(esResp.Hits.Total.Value) / float64(query.Pagination.PageSize)))
	}

	return &dto.SearchResponse{
		Results:    results,
		TotalHits:  esResp.Hits.Total.Value,
		Page:       query.Pagination.Page,
		PageSize:   query.Pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// buildQuery translates SearchQuery fields into ES DSL structure.
func (E *ElasticSearchImpl) buildQuery(query *dto.SearchQuery) map[string]interface{} {
	// Base text match
	mustClauses := []map[string]interface{}{}
	if query.Term != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":    query.Term,
				"fields":   []string{"name^3", "description", "farmer_name"},
				"fuzziness": "AUTO",
			},
		})
	}

	// Structured filters
	filterClauses := []map[string]interface{}{}
	f := query.Filters

	if f.Type != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{"type": f.Type},
		})
	}
	if f.Department != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{"department": f.Department},
		})
	}
	if f.Municipality != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{"municipality": f.Municipality},
		})
	}
	if f.FarmerID != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{"user_id": f.FarmerID},
		})
	}
	if f.PriceMin != nil || f.PriceMax != nil {
		priceRange := map[string]interface{}{}
		if f.PriceMin != nil {
			priceRange["gte"] = *f.PriceMin
		}
		if f.PriceMax != nil {
			priceRange["lte"] = *f.PriceMax
		}
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{"price": priceRange},
		})
	}

	// Bool query
	boolQuery := map[string]interface{}{}
	if len(mustClauses) > 0 {
		boolQuery["must"] = mustClauses
	} else {
		// No text term — match all
		boolQuery["must"] = []map[string]interface{}{{"match_all": map[string]interface{}{}}}
	}
	if len(filterClauses) > 0 {
		boolQuery["filter"] = filterClauses
	}

	// Score rules — application-layer relevance translated into function_score
	var queryBody map[string]interface{}
	if len(query.ScoreRules) > 0 {
		functions := make([]map[string]interface{}, 0, len(query.ScoreRules))
		for _, rule := range query.ScoreRules {
			functions = append(functions, map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{rule.Field: rule.Value},
				},
				"weight": rule.Boost,
			})
		}
		queryBody = map[string]interface{}{
			"query": map[string]interface{}{
				"function_score": map[string]interface{}{
					"query":       map[string]interface{}{"bool": boolQuery},
					"functions":   functions,
					"score_mode":  "sum",
					"boost_mode":  "multiply",
				},
			},
		}
	} else {
		queryBody = map[string]interface{}{
			"query": map[string]interface{}{"bool": boolQuery},
		}
	}

	// Sort
	if query.Sort.Field != "" && query.Sort.Field != dto.SortRelevance {
		var sortClause map[string]interface{}
		switch query.Sort.Field {
		case dto.SortPriceAsc:
			sortClause = map[string]interface{}{"price": map[string]interface{}{"order": "asc"}}
		case dto.SortPriceDesc:
			sortClause = map[string]interface{}{"price": map[string]interface{}{"order": "desc"}}
		case dto.SortProximity:
			sortClause = map[string]interface{}{
				"_geo_distance": map[string]interface{}{
					"location": map[string]interface{}{
						"lat": query.Sort.Latitude,
						"lon": query.Sort.Longitude,
					},
					"order": "asc",
					"unit":  "km",
				},
			}
		}
		if sortClause != nil {
			queryBody["sort"] = []interface{}{sortClause, "_score"}
		}
	} else {
		queryBody["sort"] = []interface{}{"_score"}
	}

	return queryBody
}

func (E *ElasticSearchImpl) Index(ctx context.Context, p *dto.IndexOfferingRequest) error {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Index: error in json marshal: %w", err)
	}

	_, err = E.client.Index(
		E.index,
		bytes.NewReader(jsonData),
		E.client.Index.WithDocumentID(p.ID),
		E.client.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("Index: error indexing document: %w", err)
	}

	return nil
}

func (E *ElasticSearchImpl) Update(ctx context.Context, id string, p *dto.IndexOfferingRequest) error {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Update: error in json marshal: %w", err)
	}

	doc := map[string]interface{}{
		"doc":           json.RawMessage(jsonData),
		"doc_as_upsert": true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("Update: error encoding update body: %w", err)
	}

	_, err = E.client.Update(
		E.index,
		id,
		&buf,
		E.client.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("Update: error updating document %s: %w", id, err)
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
		return fmt.Errorf("Delete: error deleting document %s: %w", id, err)
	}

	return nil
}
