package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type MatchUseCase interface {
	ListPrioritized(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error)
	Like(ctx context.Context, supplyOfferID uuid.UUID) (*dto.MatchDTO, *dto.TransactionDTO, error)
	Pass(ctx context.Context, supplyOfferID uuid.UUID) error
	GetByID(ctx context.Context, matchID uuid.UUID) (*dto.MatchDTO, error)
	ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.MatchDTO, error)
}
