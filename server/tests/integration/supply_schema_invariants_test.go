package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestSupplierInventoryQuantityRejectsNegative locks the 000023 invariant. The
// entity rejects a negative quantity in Go, but that left the column writable
// by anything that is not the entity. This proves the database now refuses.
func TestSupplierInventoryQuantityRejectsNegative(t *testing.T) {
	setupTransactionTestData(t)

	ctx := context.Background()
	const product = "Wheat"

	t.Run("existing rows satisfy the constraint", func(t *testing.T) {
		// The migration ran against a schema that already held data, so
		// reaching this point at all proves no row violated it.
		var count int
		if err := TestPool.QueryRow(ctx,
			`SELECT count(*) FROM supplier_inventory WHERE quantity < 0`).Scan(&count); err != nil {
			t.Fatalf("scan negative quantities: %v", err)
		}
		if count != 0 {
			t.Fatalf("supplier_inventory rows with negative quantity = %d, want 0", count)
		}
	})

	t.Run("insert below zero is rejected", func(t *testing.T) {
		_, err := TestPool.Exec(ctx,
			`INSERT INTO supplier_inventory (id, supplier_id, product_name, quantity, amount_unit)
			 VALUES ($1, $2, $3, -1, 0)`,
			testTransactionID3, testTransactionSupplierID, product)
		if err == nil {
			t.Fatal("insert of quantity = -1 succeeded, want a check violation")
		}
		assertCheckViolation(t, err, "ck_supplier_inventory_quantity")
	})

	t.Run("update below zero is rejected", func(t *testing.T) {
		id := seedInventory(t, 10, product)
		_, err := TestPool.Exec(ctx, `UPDATE supplier_inventory SET quantity = -0.5 WHERE id = $1`, id)
		if err == nil {
			t.Fatal("update to quantity = -0.5 succeeded, want a check violation")
		}
		assertCheckViolation(t, err, "ck_supplier_inventory_quantity")
	})

	t.Run("zero and positive are accepted", func(t *testing.T) {
		id := seedInventory(t, 10, product+" ok")
		if _, err := TestPool.Exec(ctx, `UPDATE supplier_inventory SET quantity = 0 WHERE id = $1`, id); err != nil {
			t.Errorf("quantity = 0 rejected: %v", err)
		}
		if _, err := TestPool.Exec(ctx, `UPDATE supplier_inventory SET quantity = 7.5 WHERE id = $1`, id); err != nil {
			t.Errorf("quantity = 7.5 rejected: %v", err)
		}
	})
}

// TestTransactionListByRequestIndexesSettleTheOpenQuestion settles a question
// this round raised: the transactions table was reported as having no indexes at
// all, which would make the aggregate in autoCloseRequest a full table read on
// every delivery confirmation. It does have two — the primary key and the UNIQUE
// on match_id, because a UNIQUE constraint creates an index in Postgres — and
// together with idx_matches_request_status on matches they cover every predicate
// ListByRequest uses. No third index is needed.
//
// This asserts the structural guarantee, not a plan shape. At fixture scale the
// planner hash-joins and sequentially scans transactions, which is the correct
// choice for a handful of rows; it switches to an index probe on match_id as the
// table grows. Pinning a plan at fixture scale would be brittle and would say
// nothing about production, so the plan is logged and the indexes are asserted.
func TestTransactionListByRequestIndexesSettleTheOpenQuestion(t *testing.T) {
	setupTransactionTestData(t)

	const listByRequest = `
		SELECT t.id, t.match_id, t.status, t.buyer_start_confirmed_at, t.supplier_start_confirmed_at,
		       t.buyer_delivery_confirmed_at, t.supplier_delivery_confirmed_at, t.cancelled_by, t.cancel_reason,
		       t.created_at, t.updated_at
		FROM transactions t
		JOIN matches m ON t.match_id = m.id
		WHERE m.supply_request_id = $1
	`

	plan := explain(t, listByRequest, testTransactionRequestID)
	t.Logf("plan for ListByRequest at fixture scale:\n%s", plan)

	// The request filter must be index-covered.
	if !strings.Contains(plan, "idx_matches_request_status") {
		t.Errorf("request filter is unindexed, idx_matches_request_status unused:\n%s", plan)
	}

	// The join side must have an index the planner can switch to.
	indexes := indexedColumns(t, "transactions")
	for _, column := range []string{"match_id", "id"} {
		if !indexes[column] {
			t.Errorf("transactions has no index covering %s, so the join degrades to a full table read. "+
				"indexed columns present: %v", column, indexes)
		}
	}

	// Capability check: with the sequential path disabled the planner must reach
	// a transactions index. This proves the index is usable, not that it is
	// always chosen. SET LOCAL is scoped to this transaction so the pool is not
	// left with sequential scans disabled for other tests.
	probe := explainInTx(t, listByRequest, []string{"SET LOCAL enable_seqscan = off"}, testTransactionRequestID)
	t.Logf("plan with sequential scans disabled:\n%s", probe)
	if !strings.Contains(probe, "transactions_match_id_key") && !strings.Contains(probe, "transactions_pkey") {
		t.Errorf("planner cannot reach any transactions index even with seqscan off:\n%s", probe)
	}
}

// explain runs EXPLAIN (FORMAT TEXT) and returns the plan lines joined.
func explain(t *testing.T, query string, args ...any) string {
	t.Helper()
	return explainInTx(t, query, nil, args...)
}

// explainInTx runs EXPLAIN inside a transaction so callers can scope planner
// settings with SET LOCAL without leaking them into the shared pool.
func explainInTx(t *testing.T, query string, setup []string, args ...any) string {
	t.Helper()

	ctx := context.Background()
	tx, err := TestPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin explain transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, statement := range setup {
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatalf("explain setup %q: %v", statement, err)
		}
	}

	rows, err := tx.Query(ctx, "EXPLAIN (FORMAT TEXT) "+query, args...)
	if err != nil {
		t.Fatalf("explain %q: %v", query, err)
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan explain row: %v", err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate explain rows: %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("EXPLAIN %q returned no plan", query)
	}
	return strings.Join(lines, "\n")
}

// indexedColumns reports which columns of the table are covered by some index.
func indexedColumns(t *testing.T, table string) map[string]bool {
	t.Helper()

	rows, err := TestPool.Query(context.Background(),
		`SELECT indexdef FROM pg_indexes WHERE tablename = $1`, table)
	if err != nil {
		t.Fatalf("read indexes for %s: %v", table, err)
	}
	defer rows.Close()

	covered := make(map[string]bool)
	for rows.Next() {
		var def string
		if err := rows.Scan(&def); err != nil {
			t.Fatalf("scan index row: %v", err)
		}
		for _, column := range []string{"match_id", "id", "status"} {
			if strings.Contains(def, "("+column) {
				covered[column] = true
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate index rows: %v", err)
	}
	return covered
}

func seedInventory(t *testing.T, quantity float64, product string) string {
	t.Helper()

	var id string
	if err := TestPool.QueryRow(context.Background(),
		`INSERT INTO supplier_inventory (supplier_id, product_name, quantity, amount_unit)
		 VALUES ($1, $2, $3, 0) RETURNING id::text`,
		testTransactionSupplierID, product, quantity).Scan(&id); err != nil {
		t.Fatalf("seed supplier_inventory: %v", err)
	}
	return id
}

func assertCheckViolation(t *testing.T, err error, wantConstraint string) {
	t.Helper()

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("error is not a *pgconn.PgError: %v", err)
	}
	if pgErr.Code != "23514" {
		t.Fatalf("SQLSTATE = %s, want 23514 (check_violation): %v", pgErr.Code, err)
	}
	if pgErr.ConstraintName != wantConstraint {
		t.Errorf("constraint = %q, want %q", pgErr.ConstraintName, wantConstraint)
	}
}
