package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	port "milpa/domain/port/secondary"
)

const rollbackTimeout = 5 * time.Second

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

		u.rollback(ctx, tx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		u.rollback(ctx, tx)
		return fmt.Errorf("unitOfWork.Commit: %w", err)
	}

	return nil
}

func (u *unitOfWork) rollback(ctx context.Context, tx pgx.Tx) {
	rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
	defer cancel()
	_ = tx.Rollback(rbCtx)
}

func newTxScope(tx pgx.Tx) port.TxScope {
	return port.TxScope{
		Requests:      &SupplyRequestRepositoryImpl{pool: tx},
		Offers:        &SupplyOfferRepositoryImpl{db: tx},
		Matches:       &MatchRepositoryImpl{db: tx},
		Transactions:  &TransactionRepositoryImpl{db: tx},
		Conversations: &ConversationRepositoyImpl{pool: tx},
	}
}
