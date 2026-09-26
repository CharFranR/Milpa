package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type SupplyRequestUseCase interface {
	Create(ctx context.Context, req dto.SupplyRequestDTO) (*dto.SupplyRequestDTO, error)
	Update(ctx context.Context, id uuid.UUID, req dto.SupplyGeneralUpdateDTO) error
	UpdateAmounts(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateAmountsDTO) error
	UpdateDeadlines(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateTimeDTO) error
	Cancel(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyRequestDTO, error)
	List(ctx context.Context) ([]*dto.SupplyRequestDTO, error)
}
