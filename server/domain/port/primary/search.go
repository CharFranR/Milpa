package primary

import (
	"context"
	"milpa/aplication/dto"
)

type FuzzyUseCase interface {
	Search(ctx context.Context, term string) ([]dto.FuzzySearchDto, error)
}
