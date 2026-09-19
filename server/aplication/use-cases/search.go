package usecases

import (
	"context"
	"fmt"
	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"
)

type SearchImpl struct {
	fuzzyRetrival port.FuzzySearch
}

func NewSearchImpl(fuzzy port.FuzzySearch) *SearchImpl {
	return &SearchImpl{
		fuzzyRetrival: fuzzy,
	}
}

func (s *SearchImpl) Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error) {

	response, err := s.fuzzyRetrival.Search(ctx, term)

	if err != nil {
		return nil, fmt.Errorf("SearchImpl Search error: %w", err)
	}

	return response, nil

}
