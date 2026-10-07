package primary

import (
	"context"

	"milpa/aplication/dto"
)

type UnitOfMeasureUseCase interface {
	List(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error)
}