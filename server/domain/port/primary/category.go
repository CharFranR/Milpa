package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type CategoryUseCase interface {
	GetAll(ctx context.Context) ([]*dto.CategoryDTO, error)
	Create(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryDTO, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateCategoryRequest) (*dto.CategoryDTO, error)
	SetStatus(ctx context.Context, id uuid.UUID, req dto.CategoryStatusRequest) (*dto.CategoryDTO, error)
}
