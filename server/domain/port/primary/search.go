package primary

import (
	"context"
	"milpa/aplication/dto"
)

type FuzzyUseCase interface {
	Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
}
