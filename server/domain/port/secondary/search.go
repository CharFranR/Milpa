package port

import (
	"context"
	"milpa/aplication/dto"
)

type FuzzyRetrival interface {
	Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error)
}

// type VectorRetrival interface {

// }
