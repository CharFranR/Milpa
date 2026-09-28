package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	port "milpa/domain/port/secondary"
)

const rollbackTimeout = 5 * time.Second

// unitOfWork opens top-level transactions on a real pool and hands the callback
// a scope whose repositories all run on that transaction.
type unitOfWork struct {
	db DB
}

func NewUnitOfWork(db DB) port.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) WithinTx(ctx context.Context, fn func(port.TxScope) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unitOfWork.Begin: %w", err)
	}

	if err := fn(newTxScope(tx)); err != nil {
		// The rollback must run even when the failure came from the request
		// context being cancelled, which is the whole point of rollback().
		u.rollback(ctx, tx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		u.rollback(ctx, tx)
		return fmt.Errorf("unitOfWork.Commit: %w", err)
	}

	return nil
}

// rollback MUST NOT use the request context: if the client disconnected, the
// ctx is already canceled and the rollback would fail, leaking the connection
// and retaining row locks. WithoutCancel preserves request-scoped values.
func (u *unitOfWork) rollback(ctx context.Context, tx pgx.Tx) {
	rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
	defer cancel()
	_ = tx.Rollback(rbCtx)
}

// newTxScope is the ONLY place a TxScope is built. Keeping the factory
// unexported is what makes the invariant "a scope's repositories all run on the
// same transaction" checkable: there is no exported way to assemble a scope
// from a mix of pool and transaction, so no repository in a scope can quietly
// escape the transaction and run standalone.
func newTxScope(tx pgx.Tx) port.TxScope {
	return port.TxScope{
		Requests:     &SupplyRequestRepositoryImpl{pool: tx},
		Offers:       &SupplyOfferRepositoryImpl{db: tx},
		Matches:      &MatchRepositoryImpl{db: tx},
		Transactions: &TransactionRepositoryImpl{db: tx},
	}
}
