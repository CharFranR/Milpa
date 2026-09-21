package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type LiquidationUseCase interface {
	CreateLiquidation(ctx context.Context, req dto.CreateLiquidationRequest) (*dto.LiquidationDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.LiquidationDTO, error)
	GetBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.LiquidationDTO, error)
	GetOpen(ctx context.Context) ([]*dto.LiquidationDTO, error)
	UpdateLiquidation(ctx context.Context, id uuid.UUID, req dto.UpdateLiquidationRequest) error
	DeleteLiquidation(ctx context.Context, id uuid.UUID) error
}
