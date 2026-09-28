package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

// ErrAmountConstraint reports that PostgreSQL refused a write because it would
// have violated ck_supply_requests_amounts
// (total_amount >= 0 AND actual_amount >= 0 AND actual_amount <= total_amount).
//
// It exists so a caller can tell "the database rejected this arithmetic" apart
// from an opaque driver failure with errors.Is, while the wrapped *pgconn.PgError
// stays in the chain for anyone who needs the constraint name PostgreSQL
// reported.
var ErrAmountConstraint = errors.New("supply request amount constraint violated")

// isCheckViolation mirrors isUniqueViolation for SQLSTATE 23514 (check_violation),
// the class of failure the amount CHECK constraint raises.
func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}

// SupplyRequestRepositoryImpl keeps a DB and not a Querier: Create and Update
// both touch addresses and supply_requests, so each of them opens its own
// top-level transaction.
//
// The narrow reservation methods below (LockForUpdate, Reserve, Release,
// UpdateStatus) are the only members a TxScope is allowed to use, and a
// tx-bound instance is built by newTxScope.
type SupplyRequestRepositoryImpl struct {
	pool DB
}

func NewSupplyRequestRepository(pool DB) *SupplyRequestRepositoryImpl {
	return &SupplyRequestRepositoryImpl{pool: pool}
}

func scanSupplyRequest(scan func(dest ...any) error) (domain.SupplyRequest, error) {
	var supplyRequest domain.SupplyRequest

	err := scan(
		&supplyRequest.ID, &supplyRequest.BuyerID, &supplyRequest.ProductName, &supplyRequest.TotalAmount,
		&supplyRequest.ActualAmount, &supplyRequest.AmountUnit, &supplyRequest.NumberOfUnits, &supplyRequest.AmountPerUnit,
		&supplyRequest.UnitOfMeasure, &supplyRequest.RequestDeadline, &supplyRequest.DeliveryDeadline, &supplyRequest.Description,
		&supplyRequest.MultipleProviders, &supplyRequest.MinAmountPerProvider, &supplyRequest.Status,
		&supplyRequest.CreatedAt, &supplyRequest.UpdatedAt,
		&supplyRequest.Address.ID, &supplyRequest.Address.Department, &supplyRequest.Address.Municipality,
		&supplyRequest.Address.AddressLine, &supplyRequest.Address.Latitude, &supplyRequest.Address.Longitude,
	)
	if err != nil {
		return domain.SupplyRequest{}, err
	}

	return supplyRequest, nil
}

// Create is TOP-LEVEL ONLY: it opens its own transaction and must NOT be called
// from inside a TxScope, where a nested Begin would only emit a SAVEPOINT.
func (r *SupplyRequestRepositoryImpl) Create(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("supplyRequest.Create: %w", err)
	}
	defer tx.Rollback(ctx)

	var addressID uuid.UUID

	if supplyRequest.Address.Department != "" {
		query := `
			INSERT INTO addresses (id, department, municipality, address_line, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		err = tx.QueryRow(ctx, query, uuid.New(), supplyRequest.Address.Department, supplyRequest.Address.Municipality,
			supplyRequest.Address.AddressLine, supplyRequest.Address.Latitude, supplyRequest.Address.Longitude).Scan(&addressID)
		if err != nil {
			return fmt.Errorf("supplyRequest.Create: %w", err)
		}
	}

	query := `
		INSERT INTO supply_requests (id, buyer_id, address_id, product_name, total_amount, actual_amount, amount_unit,
			number_of_units, amount_per_unit, unit_of_measure, request_deadline, delivery_deadline, description,
			multiple_providers, min_amount_per_provider, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err = tx.Exec(ctx, query,
		supplyRequest.ID, supplyRequest.BuyerID, nullUUID(addressID), supplyRequest.ProductName,
		supplyRequest.TotalAmount, supplyRequest.ActualAmount, supplyRequest.AmountUnit, supplyRequest.NumberOfUnits,
		supplyRequest.AmountPerUnit, supplyRequest.UnitOfMeasure, supplyRequest.RequestDeadline, supplyRequest.DeliveryDeadline,
		supplyRequest.Description, supplyRequest.MultipleProviders, supplyRequest.MinAmountPerProvider,
		supplyRequest.Status, supplyRequest.CreatedAt, supplyRequest.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("supplyRequest.Create: insert supply request: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *SupplyRequestRepositoryImpl) List(ctx context.Context, buyerID uuid.UUID) ([]domain.SupplyRequest, error) {
	query := `
		SELECT s.id, s.buyer_id, s.product_name, s.total_amount, s.actual_amount, s.amount_unit, s.number_of_units,
		       s.amount_per_unit, s.unit_of_measure, s.request_deadline, s.delivery_deadline, s.description,
		       s.multiple_providers, s.min_amount_per_provider, s.status, s.created_at, s.updated_at,
		       COALESCE(a.id, '00000000-0000-0000-0000-000000000000'), COALESCE(a.department, ''), COALESCE(a.municipality, ''),
		       COALESCE(a.address_line, ''), COALESCE(a.latitude, 0), COALESCE(a.longitude, 0)
		FROM supply_requests s
		LEFT JOIN addresses a ON s.address_id = a.id
		WHERE s.buyer_id = $1
	`

	rows, err := r.pool.Query(ctx, query, buyerID)
	if err != nil {
		return nil, fmt.Errorf("supplyRequest.List: %w", err)
	}
	defer rows.Close()

	var supplyRequests []domain.SupplyRequest
	for rows.Next() {
		supplyRequest, err := scanSupplyRequest(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("supplyRequest.List: %w", err)
		}
		supplyRequests = append(supplyRequests, supplyRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("supplyRequest.List: %w", err)
	}

	return supplyRequests, nil
}

func (r *SupplyRequestRepositoryImpl) ListOpen(ctx context.Context) ([]domain.SupplyRequest, error) {
	query := `
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

	rows, err := r.pool.Query(ctx, query, domain.SupplyRequestOpen)
	if err != nil {
		return nil, fmt.Errorf("supplyRequest.ListOpen: %w", err)
	}
	defer rows.Close()

	var supplyRequests []domain.SupplyRequest
	for rows.Next() {
		supplyRequest, err := scanSupplyRequest(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("supplyRequest.ListOpen: %w", err)
		}
		supplyRequests = append(supplyRequests, supplyRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("supplyRequest.ListOpen: %w", err)
	}

	return supplyRequests, nil
}

func (r *SupplyRequestRepositoryImpl) GetByID(ctx context.Context, supplyRequestID uuid.UUID) (domain.SupplyRequest, error) {
	query := `
		SELECT s.id, s.buyer_id, s.product_name, s.total_amount, s.actual_amount, s.amount_unit, s.number_of_units,
		       s.amount_per_unit, s.unit_of_measure, s.request_deadline, s.delivery_deadline, s.description,
		       s.multiple_providers, s.min_amount_per_provider, s.status, s.created_at, s.updated_at,
		       COALESCE(a.id, '00000000-0000-0000-0000-000000000000'), COALESCE(a.department, ''), COALESCE(a.municipality, ''),
		       COALESCE(a.address_line, ''), COALESCE(a.latitude, 0), COALESCE(a.longitude, 0)
		FROM supply_requests s
		LEFT JOIN addresses a ON s.address_id = a.id
		WHERE s.id = $1
	`

	supplyRequest, err := scanSupplyRequest(r.pool.QueryRow(ctx, query, supplyRequestID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplyRequest{}, fmt.Errorf("supplyRequest.GetByID: %w", domain.ErrNotFound)
		}
		return domain.SupplyRequest{}, fmt.Errorf("supplyRequest.GetByID: %w", err)
	}

	return supplyRequest, nil
}

// Update is TOP-LEVEL ONLY: it opens its own transaction and must NOT be called
// from inside a TxScope. It rewrites the whole row, so a concurrent reservation
// released or taken between this read and this write would be lost; the
// reservation path uses Reserve/Release/UpdateStatus instead.
func (r *SupplyRequestRepositoryImpl) Update(ctx context.Context, supplyRequest *domain.SupplyRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("supplyRequest.Update: %w", err)
	}
	defer tx.Rollback(ctx)

	if supplyRequest.Address.ID != uuid.Nil {
		query := `
			UPDATE addresses
			SET department = $1, municipality = $2, address_line = $3, latitude = $4, longitude = $5
			WHERE id = $6
		`
		_, err = tx.Exec(ctx, query,
			supplyRequest.Address.Department, supplyRequest.Address.Municipality, supplyRequest.Address.AddressLine,
			supplyRequest.Address.Latitude, supplyRequest.Address.Longitude, supplyRequest.Address.ID,
		)
		if err != nil {
			return fmt.Errorf("supplyRequest.Update: address update: %w", err)
		}
	}

	query := `
		UPDATE supply_requests
		SET product_name = $1, total_amount = $2, actual_amount = $3, amount_unit = $4, number_of_units = $5,
		    amount_per_unit = $6, unit_of_measure = $7, request_deadline = $8, delivery_deadline = $9, description = $10,
		    multiple_providers = $11, min_amount_per_provider = $12, status = $13, updated_at = $14, address_id = $15
		WHERE id = $16
	`
	_, err = tx.Exec(ctx, query,
		supplyRequest.ProductName, supplyRequest.TotalAmount, supplyRequest.ActualAmount, supplyRequest.AmountUnit,
		supplyRequest.NumberOfUnits, supplyRequest.AmountPerUnit, supplyRequest.UnitOfMeasure, supplyRequest.RequestDeadline,
		supplyRequest.DeliveryDeadline, supplyRequest.Description, supplyRequest.MultipleProviders,
		supplyRequest.MinAmountPerProvider, supplyRequest.Status, supplyRequest.UpdatedAt,
		nullUUID(supplyRequest.Address.ID), supplyRequest.ID,
	)
	if err != nil {
		return fmt.Errorf("supplyRequest.Update: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *SupplyRequestRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM supply_requests WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("supplyRequest.Delete: %w", err)
	}
	return nil
}

// LockForUpdate reads the request and holds a row lock on it for the rest of the
// transaction. It is the first lock taken by every unit of work, so no other
// transaction can observe or change this request while the unit of work runs.
//
// The lock is restricted to `s` with FOR UPDATE OF: PostgreSQL refuses
// FOR UPDATE on the nullable side of the LEFT JOIN with addresses.
func (r *SupplyRequestRepositoryImpl) LockForUpdate(ctx context.Context, id uuid.UUID) (domain.SupplyRequest, error) {
	query := `
		SELECT s.id, s.buyer_id, s.product_name, s.total_amount, s.actual_amount, s.amount_unit, s.number_of_units,
		       s.amount_per_unit, s.unit_of_measure, s.request_deadline, s.delivery_deadline, s.description,
		       s.multiple_providers, s.min_amount_per_provider, s.status, s.created_at, s.updated_at,
		       COALESCE(a.id, '00000000-0000-0000-0000-000000000000'), COALESCE(a.department, ''), COALESCE(a.municipality, ''),
		       COALESCE(a.address_line, ''), COALESCE(a.latitude, 0), COALESCE(a.longitude, 0)
		FROM supply_requests s
		LEFT JOIN addresses a ON s.address_id = a.id
		WHERE s.id = $1
		FOR UPDATE OF s
	`

	supplyRequest, err := scanSupplyRequest(r.pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplyRequest{}, fmt.Errorf("supplyRequest.LockForUpdate: %w", domain.ErrNotFound)
		}
		return domain.SupplyRequest{}, fmt.Errorf("supplyRequest.LockForUpdate: %w", err)
	}

	return supplyRequest, nil
}

// Reserve takes amount out of actual_amount only if the row still holds it. The
// predicate is the storage-level safety net: over-assignment is impossible even
// if the lock discipline is broken later, and RowsAffected == 0 is reported as
// the same domain error the state machine would have raised.
func (r *SupplyRequestRepositoryImpl) Reserve(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error {
	query := `
		UPDATE supply_requests
		SET actual_amount = actual_amount - $1, updated_at = $2
		WHERE id = $3 AND actual_amount >= $1
	`

	tag, err := r.pool.Exec(ctx, query, amount, at, id)
	if err != nil {
		return fmt.Errorf("supplyRequest.Reserve: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("supplyRequest.Reserve: %w", domain.ErrInsufficientAmount)
	}

	return nil
}

// Release gives amount back to actual_amount.
//
// It is a plain addition on purpose. The statement used to clamp with
// LEAST(actual_amount + $1, total_amount), which meant a double release was an
// invisible no-op: the write succeeded, no error surfaced anywhere, and the only
// symptom was an actual_amount that quietly disagreed with the transaction log.
// Once ck_supply_requests_amounts exists, the clamp is redundant for every
// legitimate path (Reserve is guarded by actual_amount >= amount and the request
// row is locked for the whole unit of work) and actively harmful for the
// illegitimate one, so the addition is left bare and the CHECK constraint is
// allowed to reject an over-release as the corruption it is.
func (r *SupplyRequestRepositoryImpl) Release(ctx context.Context, id uuid.UUID, amount float64, at time.Time) error {
	query := `
		UPDATE supply_requests
		SET actual_amount = actual_amount + $1, updated_at = $2
		WHERE id = $3
	`

	if _, err := r.pool.Exec(ctx, query, amount, at, id); err != nil {
		if isCheckViolation(err) {
			return fmt.Errorf("supplyRequest.Release: releasing %v would push actual_amount past total_amount: %w: %w",
				amount, ErrAmountConstraint, err)
		}
		return fmt.Errorf("supplyRequest.Release: %w", err)
	}

	return nil
}

// UpdateStatus writes only status and updated_at. Writing the whole row here
// would reintroduce the lost update this refactor removes: the unit of work
// holds a snapshot of actual_amount from the start of its transaction, and a
// full-row write would push that snapshot back over any concurrent change.
func (r *SupplyRequestRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error {
	query := `
		UPDATE supply_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	if _, err := r.pool.Exec(ctx, query, status, at, id); err != nil {
		return fmt.Errorf("supplyRequest.UpdateStatus: %w", err)
	}

	return nil
}

// UpdateCompletion closes out a fulfilled request by writing status, zeroing
// actual_amount and stamping updated_at in a single statement.
//
// It stays narrow on purpose, naming the three columns it owns. A full-row
// write here would reintroduce the lost update this refactor exists to remove:
// the caller is inside a unit of work that read actual_amount at the start of
// its transaction, and pushing that snapshot back over the row would discard
// any concurrent change. Status plus the zeroing is all completion requires.
//
// The zeroing is a real semantic statement, not cosmetics. Accumulating
// fractional reservations in a DOUBLE PRECISION column leaves an IEEE-754
// residue, so actual_amount drifts to 2.220446049250313e-16 for a 2.7 request
// filled by three 0.9 deliveries and can never land on an exact zero by
// subtraction. A completed request has nothing left to fulfil, so the column
// is set rather than approximated.
func (r *SupplyRequestRepositoryImpl) UpdateCompletion(ctx context.Context, id uuid.UUID, status domain.SupplyRequestStatus, at time.Time) error {
	query := `
		UPDATE supply_requests
		SET status = $1, actual_amount = 0, updated_at = $2
		WHERE id = $3
	`

	if _, err := r.pool.Exec(ctx, query, status, at, id); err != nil {
		return fmt.Errorf("supplyRequest.UpdateCompletion: %w", err)
	}

	return nil
}

var (
	_ port.SupplyRequestRepository = (*SupplyRequestRepositoryImpl)(nil)
	_ port.RequestReservationStore = (*SupplyRequestRepositoryImpl)(nil)
)
