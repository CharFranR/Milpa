package integration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

const openRequestIndexName = "idx_supply_requests_status_created"

const listOpenPlanQuery = `
	SELECT s.id, s.buyer_id, s.product_name, s.total_amount, s.actual_amount, s.amount_unit, s.number_of_units,
	       s.amount_per_unit, s.unit_of_measure, s.request_deadline, s.delivery_deadline, s.description,
	       s.multiple_providers, s.min_amount_per_provider, s.status, s.created_at, s.updated_at,
	       COALESCE(a.id, '00000000-0000-0000-0000-000000000000'), COALESCE(a.department, ''), COALESCE(a.municipality, ''),
	       COALESCE(a.address_line, ''), COALESCE(a.latitude, 0), COALESCE(a.longitude, 0)
	FROM supply_requests s
	LEFT JOIN addresses a ON s.address_id = a.id
	WHERE s.status = $1
	ORDER BY s.created_at DESC
`

type explainNode struct {
	NodeType     string        `json:"Node Type"`
	IndexName    string        `json:"Index Name"`
	RelationName string        `json:"Relation Name"`
	Plans        []explainNode `json:"Plans"`
}

type explainResult struct {
	Plan explainNode `json:"Plan"`
}

func walkPlan(node explainNode, visit func(explainNode)) {
	visit(node)
	for _, child := range node.Plans {
		walkPlan(child, visit)
	}
}

// TestOpenRequestIndexExistsWithTheExpectedShape asserts the index from migration
// 000022 is present, leading with status and ordering created_at descending.
//
// The migration is also replayed by the TestContainers harness, but that harness
// reads files from disk, so a migration that never shipped in the embed, or an
// index built on the wrong columns, would only ever show up here.
func TestOpenRequestIndexExistsWithTheExpectedShape(t *testing.T) {
	ctx := context.Background()

	var indexDef string
	err := TestPool.QueryRow(ctx, `SELECT indexdef FROM pg_indexes WHERE indexname = $1`, openRequestIndexName).Scan(&indexDef)
	if errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("index %s does not exist: it is created by migration 000022, which the harness replays from disk",
			openRequestIndexName)
	}
	if err != nil {
		t.Fatalf("read the definition of %s: %v", openRequestIndexName, err)
	}

	// Column order and per-column sort direction come from the catalog rather
	// than from parsing indexdef text, so the assertion survives formatting
	// changes between PostgreSQL versions. indoption holds two bits per key: the
	// low bit set means that key is DESC, so the whole comparison stays in SQL
	// instead of decoding a bit vector in Go.
	rows, err := TestPool.Query(ctx, `
		SELECT a.attname, (i.indoption[k.ord - 1] & 1) = 1
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		CROSS JOIN LATERAL unnest(i.indkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum
		WHERE c.relname = $1
		ORDER BY k.ord
	`, openRequestIndexName)
	if err != nil {
		t.Fatalf("inspect the columns of %s: %v", openRequestIndexName, err)
	}
	defer rows.Close()

	type column struct {
		name       string
		descending bool
	}
	var columns []column
	for rows.Next() {
		var name string
		var descending bool
		if err := rows.Scan(&name, &descending); err != nil {
			t.Fatalf("scan a column of %s: %v", openRequestIndexName, err)
		}
		columns = append(columns, column{name: name, descending: descending})
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate the columns of %s: %v", openRequestIndexName, err)
	}

	if len(columns) != 2 {
		t.Fatalf("index %s has %d key columns (%v), want exactly 2: %s",
			openRequestIndexName, len(columns), columns, indexDef)
	}
	if columns[0].name != "status" {
		t.Errorf("index %s leads with %q, want status: a global status scan cannot use an index whose leading column is buyer_id",
			openRequestIndexName, columns[0].name)
	}
	if columns[0].descending {
		t.Errorf("index %s sorts status descending, want ascending", openRequestIndexName)
	}
	if columns[1].name != "created_at" {
		t.Errorf("index %s second key is %q, want created_at, so the ORDER BY is served by the index",
			openRequestIndexName, columns[1].name)
	}
	if !columns[1].descending {
		t.Errorf("index %s sorts created_at ascending, want descending to match ORDER BY created_at DESC",
			openRequestIndexName)
	}
}

// TestOpenRequestIndexIsAUsablePathForListOpen asserts the planner can build a
// path over the new index for the ListOpen statement, which is the property that
// makes the index worth having.
//
// enable_seqscan is turned off for the duration of the EXPLAIN. That is not
// coercion of the answer: it removes the cost model from the question, so what is
// being asked is "can this index serve this query at all", not "would the cost
// model pick it on a table with a handful of fixture rows". If the index were
// missing, malformed, or built on the wrong columns, no such path could be built
// with sequential scans disabled and the assertion would still fail. Asserting
// plan choice without this would couple the test to PostgreSQL's costing and
// flake the moment a fixture row changed.
func TestOpenRequestIndexIsAUsablePathForListOpen(t *testing.T) {
	ctx := context.Background()

	// A transaction, because SET LOCAL is scoped to one and must not leak into
	// the rest of the suite.
	tx, err := TestPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SET LOCAL enable_seqscan = off`); err != nil {
		t.Fatalf("SET LOCAL enable_seqscan = off: %v", err)
	}

	var raw []byte
	if err := tx.QueryRow(ctx, `EXPLAIN (FORMAT JSON) `+listOpenPlanQuery, domain.SupplyRequestOpen).Scan(&raw); err != nil {
		t.Fatalf("EXPLAIN the ListOpen statement: %v", err)
	}

	var plans []explainResult
	if err := json.Unmarshal(raw, &plans); err != nil {
		t.Fatalf("decode the ListOpen plan %s: %v", raw, err)
	}
	if len(plans) != 1 {
		t.Fatalf("EXPLAIN returned %d plans, want 1", len(plans))
	}

	usedIndex := false
	var seqScans []string
	walkPlan(plans[0].Plan, func(node explainNode) {
		if node.IndexName == openRequestIndexName {
			usedIndex = true
		}
		if node.NodeType == "Seq Scan" && node.RelationName == "supply_requests" {
			seqScans = append(seqScans, node.RelationName)
		}
	})

	if !usedIndex {
		t.Errorf("the ListOpen plan does not use %s even with sequential scans disabled; plan: %s",
			openRequestIndexName, raw)
	}
	if len(seqScans) > 0 {
		t.Errorf("the ListOpen plan still scans supply_requests sequentially: %v", seqScans)
	}
}

// TestListOpenReturnsNewestFirst pins the ordering contract of ListOpen against
// real data, independently of the index.
//
// created_at is set explicitly rather than inherited from time.Now() in the
// fixture so the ordering is decided by the test and cannot depend on how fast
// four inserts land relative to each other.
func TestListOpenReturnsNewestFirst(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)
	ctx := context.Background()

	// Oldest first, so an implementation that returned insertion order would
	// fail exactly as one that returned ascending order would.
	created := []time.Time{
		fixedTime.Add(1 * time.Hour),
		fixedTime.Add(2 * time.Hour),
		fixedTime.Add(3 * time.Hour),
		fixedTime.Add(4 * time.Hour),
	}

	var inserted []uuid.UUID
	for i, at := range created {
		request := f.newConcRequest(t, 10, true)
		if _, err := TestPool.Exec(ctx,
			`UPDATE supply_requests SET created_at = $1 WHERE id = $2`, at, request.ID); err != nil {
			t.Fatalf("pin created_at of request %d: %v", i, err)
		}
		inserted = append(inserted, request.ID)
	}

	// One closed request that must never appear: it is the newest row overall,
	// so an implementation that ignored the status filter would still be caught.
	closed := f.newConcRequest(t, 10, true)
	if _, err := TestPool.Exec(ctx,
		`UPDATE supply_requests SET status = $1, created_at = $2 WHERE id = $3`,
		domain.SupplyRequestCancelled, fixedTime.Add(5*time.Hour), closed.ID); err != nil {
		t.Fatalf("insert the closed request: %v", err)
	}

	open, err := f.requestRepo.ListOpen(ctx)
	if err != nil {
		t.Fatalf("ListOpen(): %v", err)
	}
	if len(open) != len(inserted) {
		t.Fatalf("ListOpen returned %d requests, want %d (one closed request must be filtered out)", len(open), len(inserted))
	}

	for i, request := range open {
		want := inserted[len(inserted)-1-i]
		if request.ID != want {
			t.Errorf("ListOpen position %d = %s, want %s: results must be ordered by created_at DESC, not by insertion",
				i, request.ID, want)
		}
	}

	for i := 1; i < len(open); i++ {
		if open[i].CreatedAt.After(open[i-1].CreatedAt) {
			t.Errorf("ListOpen is not sorted newest first at position %d: %v came after %v",
				i, open[i].CreatedAt, open[i-1].CreatedAt)
		}
	}
}

// TestReleaseRejectsOverReleaseAndLeavesTheRowUnchanged is the regression test for
// the LEAST clamp that Release used to wrap the addition in.
//
// The sequence below is a real double release: reserve the whole request, release
// it, then release the same amount again. With the clamp the third statement
// succeeded, wrote nothing, and returned no error, so the corruption was invisible
// at every layer. Now it must be refused by ck_supply_requests_amounts and the row
// must be exactly as it was.
func TestReleaseRejectsOverReleaseAndLeavesTheRowUnchanged(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	ctx := context.Background()
	const total = 100.0

	request := f.newConcRequest(t, total, true)
	now := f.clock.Now()

	if err := f.requestRepo.Reserve(ctx, request.ID, total, now); err != nil {
		t.Fatalf("Reserve(%v): %v", total, err)
	}
	if got := concActualAmount(t, request.ID); got != 0 {
		t.Fatalf("actual amount after reserving everything = %v, want 0", got)
	}

	// The legitimate release: it puts the request back exactly at its total, which
	// the CHECK constraint allows because it tests <= and not <.
	if err := f.requestRepo.Release(ctx, request.ID, total, now); err != nil {
		t.Fatalf("Release(%v) undoing the reservation: %v", total, err)
	}
	if got := concActualAmount(t, request.ID); got != total {
		t.Fatalf("actual amount after the legitimate release = %v, want %v", got, total)
	}

	// The double release. This is the statement the clamp used to swallow.
	err := f.requestRepo.Release(ctx, request.ID, total, now)
	if err == nil {
		t.Fatalf("releasing %v a second time returned no error: the LEAST clamp is still hiding the over-release", total)
	}
	if !errors.Is(err, repository.ErrAmountConstraint) {
		t.Errorf("over-release error = %v, want one wrapping repository.ErrAmountConstraint so callers can recognise it with errors.Is", err)
	}

	// The mapping must stay diagnosable: the driver error is preserved in the
	// chain, so the constraint name PostgreSQL reported is still reachable.
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("over-release error = %v, want the original *pgconn.PgError in the chain", err)
	}
	if pgErr.Code != "23514" {
		t.Errorf("SQLSTATE = %s, want 23514 (check_violation)", pgErr.Code)
	}
	if pgErr.ConstraintName != "ck_supply_requests_amounts" {
		t.Errorf("violated constraint = %q, want ck_supply_requests_amounts", pgErr.ConstraintName)
	}

	if got := concActualAmount(t, request.ID); got != total {
		t.Errorf("actual amount after the rejected over-release = %v, want %v: a refused statement must leave the row untouched",
			got, total)
	}

	// The request is still usable afterwards: the rejection was an error, not
	// corruption.
	if _, err := f.requestRepo.ListOpen(ctx); err != nil {
		t.Errorf("ListOpen() after the rejected over-release: %v", err)
	}
}
