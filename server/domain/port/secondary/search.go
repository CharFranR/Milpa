package port

import (
	"context"
	"milpa/aplication/dto"
)

type FuzzyRetrival interface {
	Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error)
	Index(ctx context.Context, p *dto.IndexOfferingRequest) error
	Update(ctx context.Context, id string, p *dto.IndexOfferingRequest) error
	Delete(ctx context.Context, id string) error
}

type FuzzySearch interface {
	Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error)
}

// type VectorRetrival interface {

// }
