package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type TransactionUseCase interface {
	GetByMatch(ctx context.Context, matchID uuid.UUID) (*dto.TransactionDTO, error)
	ConfirmStart(ctx context.Context, transactionID uuid.UUID) error
	ConfirmDelivery(ctx context.Context, transactionID uuid.UUID) error
	Cancel(ctx context.Context, transactionID uuid.UUID, reason string) error
	ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.TransactionDTO, error)
}
