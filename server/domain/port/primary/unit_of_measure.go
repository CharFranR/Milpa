package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type UnitOfMeasureUseCase interface {
	List(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error)
	ListAll(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error)
	Create(ctx context.Context, req dto.CreateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error)
	SetStatus(ctx context.Context, id uuid.UUID, req dto.UnitOfMeasureStatusRequest) (*dto.UnitOfMeasureDTO, error)
}
