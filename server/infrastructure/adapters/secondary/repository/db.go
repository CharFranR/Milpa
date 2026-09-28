package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is the surface shared by a connection pool and an open transaction.
//
// NOTE: pgx.Tx satisfies this. It does NOT satisfy DB, which is intentional.
// pgx.Tx.Begin emits a SAVEPOINT instead of a real nested transaction, so a
// repository that needs a top-level transaction must depend on DB while a
// repository that only runs statements must depend on Querier. Keeping the two
// apart means a pool can never be mistaken for a transaction at compile time
// and a transaction can never silently open a savepoint.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// DB can open a TOP-LEVEL transaction. Only *pgxpool.Pool implements this.
type DB interface {
	Querier
	Begin(ctx context.Context) (pgx.Tx, error)
}
