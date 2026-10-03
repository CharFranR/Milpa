package primary

import (
	"context"

	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type SupplierInventoryUseCase interface {
	Upsert(ctx context.Context, supplierID uuid.UUID, req dto.UpsertSupplierInventoryRequest) (*dto.SupplierInventoryDTO, error)
	ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplierInventoryDTO, error)
	Delete(ctx context.Context, supplierID uuid.UUID, id uuid.UUID) error
}
