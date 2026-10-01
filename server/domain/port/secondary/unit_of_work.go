package port

import (
	"context"
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type RequestReservationStore interface {
	LockForUpdate(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error)

	Reserve(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error

	Release(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error

	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error

	UpdateCompletion(ctx context.Context, id uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error
}

type TxOfferRepository interface {
	SupplyOfferRepository
	LockByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error)
}

type TxMatchRepository interface {
	MatchRepository
	LockByIDForUpdate(ctx context.Context, id uuid.UUID) (*domain.Match, error)
}

type TxTransactionRepository interface {
	TransactionRepository
	LockByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.Transaction, error)
}

type TxScope struct {
	Requests     RequestReservationStore
	Offers       TxOfferRepository
	Matches      TxMatchRepository
	Transactions TxTransactionRepository
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(TxScope) error) error
}
