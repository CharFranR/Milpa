package port

import (
	"context"
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

// RequestReservationStore is the narrow write surface a supply request exposes
// to a unit of work. It is deliberately NOT part of SupplyRequestRepository:
// the full-row Create/Update of that repository open their own top-level
// transaction and must never be reached from inside a TxScope.
//
// The narrowness is the point. Reserve and Release are the only statements that
// move actual_amount, and they move it with a SQL-level guard instead of a
// full-row rewrite, so a lost update cannot reintroduce a stale actual_amount.
type RequestReservationStore interface {
	// LockForUpdate reads the request and holds a row lock on it until the
	// enclosing transaction ends. It is the FIRST lock of the global order.
	LockForUpdate(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error)
	// Reserve subtracts amount from actual_amount, but only while the row holds
	// at least that much. It returns domain.ErrInsufficientAmount when the
	// statement affects no row.
	Reserve(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error
	// Release adds amount back to actual_amount, clamped to total_amount.
	Release(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error
	// UpdateStatus writes only status and updated_at. A full-row rewrite would
	// overwrite concurrent actual_amount changes with the value this unit of
	// work happened to read, so the auto-close path must not use it.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error
}

// TxOfferRepository, TxMatchRepository and TxTransactionRepository widen the
// corresponding read/write ports with the row locks that only make sense while a
// transaction is open. They live here and not in repository.go so that every
// existing fake of the read/write ports keeps compiling: a repository used
// outside a unit of work has nothing to lock against.
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

// TxScope is the set of repositories bound to a single open transaction. It is
// only ever produced by the secondary unit-of-work adapter, which is what
// guarantees that every statement a use case runs on a scope either commits
// together or rolls back together.
type TxScope struct {
	Requests     RequestReservationStore
	Offers       TxOfferRepository
	Matches      TxMatchRepository
	Transactions TxTransactionRepository
}

// UnitOfWork runs a unit of work inside one top-level transaction.
//
// Global lock order for every implementation: SupplyRequest -> SupplyOffer ->
// Match -> Transaction. Taking them in any other order invites a deadlock, and
// no implementation may deviate from it.
type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(TxScope) error) error
}
