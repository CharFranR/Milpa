package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type SupplyOfferRepositoryImpl struct {
	db Querier
}

func NewSupplyOfferRepository(pool DB) *SupplyOfferRepositoryImpl {
	return &SupplyOfferRepositoryImpl{db: pool}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// supplyOfferColumns is the single source of truth for the projection shared by
// every supply_offers read. The five read paths used to spell the column list out
// individually, which is exactly how a column ends up missing from one of them:
// ranking and matching would then disagree about the same row.
const supplyOfferColumns = `id, supplier_id, supply_request_id, total_amount, amount_unit, price_per_unit, comments,
	proposed_delivery_day, delivery_available, status, created_at, updated_at`

func scanSupplyOffer(scan func(dest ...any) error) (domain.SupplyOffer, error) {
	var supplyOffer domain.SupplyOffer

	err := scan(
		&supplyOffer.ID, &supplyOffer.SupplierID, &supplyOffer.SupplyRequest, &supplyOffer.TotalAmount,
		&supplyOffer.AmountUnit, &supplyOffer.PricePerUnit, &supplyOffer.Comments,
		&supplyOffer.ProposedDeliveryDay, &supplyOffer.DeliveryAvailable, &supplyOffer.Status,
		&supplyOffer.CreatedAt, &supplyOffer.UpdatedAt,
	)
	if err != nil {
		return domain.SupplyOffer{}, err
	}

	return supplyOffer, nil
}

func (r *SupplyOfferRepositoryImpl) Create(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	query := `
		INSERT INTO supply_offers (id, supplier_id, supply_request_id, total_amount, amount_unit,
			price_per_unit, comments, proposed_delivery_day, delivery_available, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		supplyOffer.ID, supplyOffer.SupplierID, supplyOffer.SupplyRequest, supplyOffer.TotalAmount, supplyOffer.AmountUnit,
		supplyOffer.PricePerUnit, supplyOffer.Comments,
		supplyOffer.ProposedDeliveryDay, supplyOffer.DeliveryAvailable, supplyOffer.Status,
		supplyOffer.CreatedAt, supplyOffer.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("supplyOffer.Create: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("supplyOffer.Create: %w", err)
	}
	return nil
}

func (r *SupplyOfferRepositoryImpl) List(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplyOffer, error) {
	query := `
		SELECT ` + supplyOfferColumns + `
		FROM supply_offers
		WHERE supplier_id = $1
	`

	rows, err := r.db.Query(ctx, query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("supplyOffer.List: %w", err)
	}
	defer rows.Close()

	var supplyOffers []domain.SupplyOffer
	for rows.Next() {
		supplyOffer, err := scanSupplyOffer(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("supplyOffer.List: %w", err)
		}
		supplyOffers = append(supplyOffers, supplyOffer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("supplyOffer.List: %w", err)
	}

	return supplyOffers, nil
}

func (r *SupplyOfferRepositoryImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.SupplyOffer, error) {
	query := `
		SELECT ` + supplyOfferColumns + `
		FROM supply_offers
		WHERE supply_request_id = $1
	`

	rows, err := r.db.Query(ctx, query, supplyRequestID)
	if err != nil {
		return nil, fmt.Errorf("supplyOffer.ListByRequest: %w", err)
	}
	defer rows.Close()

	var supplyOffers []domain.SupplyOffer
	for rows.Next() {
		supplyOffer, err := scanSupplyOffer(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("supplyOffer.ListByRequest: %w", err)
		}
		supplyOffers = append(supplyOffers, supplyOffer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("supplyOffer.ListByRequest: %w", err)
	}

	return supplyOffers, nil
}

func (r *SupplyOfferRepositoryImpl) FindBySupplierAndRequest(ctx context.Context, supplierID, supplyRequestID uuid.UUID) (domain.SupplyOffer, error) {
	query := `
		SELECT ` + supplyOfferColumns + `
		FROM supply_offers
		WHERE supplier_id = $1 AND supply_request_id = $2
	`

	supplyOffer, err := scanSupplyOffer(r.db.QueryRow(ctx, query, supplierID, supplyRequestID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.FindBySupplierAndRequest: %w", domain.ErrNotFound)
		}
		return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.FindBySupplierAndRequest: %w", err)
	}

	return supplyOffer, nil
}

func (r *SupplyOfferRepositoryImpl) GetByID(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
	query := `
		SELECT ` + supplyOfferColumns + `
		FROM supply_offers
		WHERE id = $1
	`

	supplyOffer, err := scanSupplyOffer(r.db.QueryRow(ctx, query, supplyOfferID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.GetByID: %w", domain.ErrNotFound)
		}
		return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.GetByID: %w", err)
	}

	return supplyOffer, nil
}

// LockByIDForUpdate reads the offer and holds a row lock on it until the
// enclosing transaction ends. It is the second lock of the global order:
// SupplyRequest -> SupplyOffer.
//
// It projects the same columns as the unlocked reads, price included, so a
// price read under the row lock is the locked price.
func (r *SupplyOfferRepositoryImpl) LockByIDForUpdate(ctx context.Context, supplyOfferID uuid.UUID) (domain.SupplyOffer, error) {
	query := `
		SELECT ` + supplyOfferColumns + `
		FROM supply_offers
		WHERE id = $1
		FOR UPDATE
	`

	supplyOffer, err := scanSupplyOffer(r.db.QueryRow(ctx, query, supplyOfferID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.LockByIDForUpdate: %w", domain.ErrNotFound)
		}
		return domain.SupplyOffer{}, fmt.Errorf("supplyOffer.LockByIDForUpdate: %w", err)
	}

	return supplyOffer, nil
}

func (r *SupplyOfferRepositoryImpl) Update(ctx context.Context, supplyOffer *domain.SupplyOffer) error {
	query := `
		UPDATE supply_offers
		SET total_amount = $1, amount_unit = $2, price_per_unit = $3, comments = $4, proposed_delivery_day = $5,
		    delivery_available = $6, status = $7, updated_at = $8
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query,
		supplyOffer.TotalAmount, supplyOffer.AmountUnit, supplyOffer.PricePerUnit, supplyOffer.Comments,
		supplyOffer.ProposedDeliveryDay, supplyOffer.DeliveryAvailable,
		supplyOffer.Status, supplyOffer.UpdatedAt, supplyOffer.ID,
	)
	if err != nil {
		return fmt.Errorf("supplyOffer.Update: %w", err)
	}
	return nil
}

func (r *SupplyOfferRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM supply_offers WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("supplyOffer.Delete: %w", err)
	}
	return nil
}

var (
	_ port.SupplyOfferRepository = (*SupplyOfferRepositoryImpl)(nil)
	_ port.TxOfferRepository     = (*SupplyOfferRepositoryImpl)(nil)
)
