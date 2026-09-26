package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type TransactionRepositoryImpl struct {
	pool DB
}

func NewTransactionRepository(pool DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{pool: pool}
}

func scanTransaction(scan func(dest ...any) error) (domain.Transaction, error) {
	var transaction domain.Transaction

	err := scan(
		&transaction.ID, &transaction.MatchID, &transaction.Status,
		&transaction.BuyerStartConfirmedAt, &transaction.SupplierStartConfirmedAt,
		&transaction.BuyerDeliveryConfirmedAt, &transaction.SupplierDeliveryConfirmedAt,
		&transaction.CancelledBy, &transaction.CancelReason, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, match_id, status, buyer_start_confirmed_at, supplier_start_confirmed_at,
			buyer_delivery_confirmed_at, supplier_delivery_confirmed_at, cancelled_by, cancel_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		transaction.ID, transaction.MatchID, transaction.Status,
		transaction.BuyerStartConfirmedAt, transaction.SupplierStartConfirmedAt,
		transaction.BuyerDeliveryConfirmedAt, transaction.SupplierDeliveryConfirmedAt,
		transaction.CancelledBy, transaction.CancelReason, transaction.CreatedAt, transaction.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("transaction.Create: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("transaction.Create: %w", err)
	}
	return nil
}

func (r *TransactionRepositoryImpl) List(ctx context.Context, matchID uuid.UUID) ([]domain.Transaction, error) {
	query := `
		SELECT id, match_id, status, buyer_start_confirmed_at, supplier_start_confirmed_at,
		       buyer_delivery_confirmed_at, supplier_delivery_confirmed_at, cancelled_by, cancel_reason, created_at, updated_at
		FROM transactions
		WHERE match_id = $1
	`

	rows, err := r.pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("transaction.List: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("transaction.List: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction.List: %w", err)
	}

	return transactions, nil
}

func (r *TransactionRepositoryImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Transaction, error) {
	query := `
		SELECT t.id, t.match_id, t.status, t.buyer_start_confirmed_at, t.supplier_start_confirmed_at,
		       t.buyer_delivery_confirmed_at, t.supplier_delivery_confirmed_at, t.cancelled_by, t.cancel_reason,
		       t.created_at, t.updated_at
		FROM transactions t
		JOIN matches m ON t.match_id = m.id
		WHERE m.supply_request_id = $1
	`

	rows, err := r.pool.Query(ctx, query, supplyRequestID)
	if err != nil {
		return nil, fmt.Errorf("transaction.ListByRequest: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("transaction.ListByRequest: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction.ListByRequest: %w", err)
	}

	return transactions, nil
}

func (r *TransactionRepositoryImpl) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Transaction, error) {
	query := `
		SELECT t.id, t.match_id, t.status, t.buyer_start_confirmed_at, t.supplier_start_confirmed_at,
		       t.buyer_delivery_confirmed_at, t.supplier_delivery_confirmed_at, t.cancelled_by, t.cancel_reason,
		       t.created_at, t.updated_at
		FROM transactions t
		JOIN matches m ON t.match_id = m.id
		JOIN supply_offers o ON m.supply_offer_id = o.id
		WHERE o.supplier_id = $1 AND t.status IN ($2, $3)
	`

	rows, err := r.pool.Query(ctx, query, supplierID, domain.TransactionMatched, domain.TransactionInProgress)
	if err != nil {
		return nil, fmt.Errorf("transaction.ListActiveBySupplier: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("transaction.ListActiveBySupplier: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction.ListActiveBySupplier: %w", err)
	}

	return transactions, nil
}

func (r *TransactionRepositoryImpl) GetByMatch(ctx context.Context, matchID uuid.UUID) (domain.Transaction, error) {
	query := `
		SELECT id, match_id, status, buyer_start_confirmed_at, supplier_start_confirmed_at,
		       buyer_delivery_confirmed_at, supplier_delivery_confirmed_at, cancelled_by, cancel_reason, created_at, updated_at
		FROM transactions
		WHERE match_id = $1
	`

	transaction, err := scanTransaction(r.pool.QueryRow(ctx, query, matchID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, fmt.Errorf("transaction.GetByMatch: %w", domain.ErrNotFound)
		}
		return domain.Transaction{}, fmt.Errorf("transaction.GetByMatch: %w", err)
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) GetByID(ctx context.Context, transactionID uuid.UUID) (domain.Transaction, error) {
	query := `
		SELECT id, match_id, status, buyer_start_confirmed_at, supplier_start_confirmed_at,
		       buyer_delivery_confirmed_at, supplier_delivery_confirmed_at, cancelled_by, cancel_reason, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	transaction, err := scanTransaction(r.pool.QueryRow(ctx, query, transactionID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, fmt.Errorf("transaction.GetByID: %w", domain.ErrNotFound)
		}
		return domain.Transaction{}, fmt.Errorf("transaction.GetByID: %w", err)
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET status = $1, buyer_start_confirmed_at = $2, supplier_start_confirmed_at = $3,
		    buyer_delivery_confirmed_at = $4, supplier_delivery_confirmed_at = $5, cancelled_by = $6,
		    cancel_reason = $7, updated_at = $8
		WHERE id = $9
	`
	_, err := r.pool.Exec(ctx, query,
		transaction.Status, transaction.BuyerStartConfirmedAt, transaction.SupplierStartConfirmedAt,
		transaction.BuyerDeliveryConfirmedAt, transaction.SupplierDeliveryConfirmedAt, transaction.CancelledBy,
		transaction.CancelReason, transaction.UpdatedAt, transaction.ID,
	)
	if err != nil {
		return fmt.Errorf("transaction.Update: %w", err)
	}
	return nil
}

func (r *TransactionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("transaction.Delete: %w", err)
	}
	return nil
}

var _ port.TransactionRepository = (*TransactionRepositoryImpl)(nil)
