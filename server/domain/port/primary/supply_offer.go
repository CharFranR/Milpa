package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type SupplyOfferUseCase interface {
	Create(ctx context.Context, req dto.SupplyOfferDTO) (*dto.SupplyOfferDTO, error)
	Update(ctx context.Context, id uuid.UUID, req dto.SupplyOfferUpdateDTO) error
	Withdraw(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyOfferDTO, error)
	ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.SupplyOfferDTO, error)
	ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyOfferDTO, error)
}
