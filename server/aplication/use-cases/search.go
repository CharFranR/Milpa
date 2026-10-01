package usecases

import (
	"context"
	"fmt"
	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type SearchImpl struct {
	fuzzyRetrival port.FuzzySearch
}

func NewSearchImpl(fuzzy port.FuzzySearch) *SearchImpl {
	return &SearchImpl{
		fuzzyRetrival: fuzzy,
	}
}

func (s *SearchImpl) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	query := s.buildQuery(req)

	response, err := s.fuzzyRetrival.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SearchImpl.Search: %w", err)
	}

	return response, nil
}

func (s *SearchImpl) buildQuery(req dto.SearchRequest) *dto.SearchQuery {
	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	query := &dto.SearchQuery{
		Term: req.Term,
		Filters: dto.SearchFilters{
			Type:         req.Type,
			CategoryID:   req.CategoryID,
			Department:   req.Department,
			Municipality: req.Municipality,
			PriceMin:     req.PriceMin,
			PriceMax:     req.PriceMax,
			FarmerID:     req.FarmerID,
		},
		Sort: dto.SearchSort{
			Field:     req.SortBy,
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		Pagination: dto.SearchPagination{
			Page:     page,
			PageSize: pageSize,
		},
	}

	query.ScoreRules = append(query.ScoreRules, dto.ScoreRule{
		Field: "farmer_verified",
		Value: true,
		Boost: 1.5,
	})

	return query
}
