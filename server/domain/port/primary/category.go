package primary

import (
	"context"

	"milpa/aplication/dto"
)

type CategoryUseCase interface {
	GetAll(ctx context.Context) ([]*dto.CategoryDTO, error)
}
