package primary

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

var ErrActiveMatch = errors.New("supply request already has an active match")

// ?

type SupplyRequestUseCase interface {
	Create(ctx context.Context, req dto.SupplyRequestDTO) (*dto.SupplyRequestDTO, error)
	Update(ctx context.Context, id uuid.UUID, req dto.SupplyGeneralUpdateDTO) error
	UpdateAmounts(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateAmountsDTO) error
	UpdateDeadlines(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateTimeDTO) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Expire(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyRequestDTO, error)
	List(ctx context.Context) ([]*dto.SupplyRequestDTO, error)
	ListAvailable(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyRequestDTO, error)
}
