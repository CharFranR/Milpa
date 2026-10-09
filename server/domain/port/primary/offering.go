package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type OfferingUseCase interface {
	CreateOffering(ctx context.Context, req dto.CreateOfferingRequest) (*dto.OfferingDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, includeHidden bool) ([]*dto.OfferingDTO, error)
	UpdateOffering(ctx context.Context, id uuid.UUID, req dto.UpdateOfferingRequest) error
	DeleteOffering(ctx context.Context, id uuid.UUID) error
	DeactivateOffering(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error)
	RenewOffering(ctx context.Context, id uuid.UUID, req dto.RenewOfferingRequest) (*dto.OfferingDTO, error)
}
