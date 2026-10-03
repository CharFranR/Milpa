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

type LiquidationRepositoryImpl struct {
	pool DB
}

func NewLiquidationRepository(pool DB) *LiquidationRepositoryImpl {
	return &LiquidationRepositoryImpl{pool: pool}
}

// liquidationColumns is the projection every liquidation read shares.
const liquidationColumns = `
	SELECT id, supplier_id, product_name, quantity, unit_of_measure,
	       total_price, unit_price, delivery_time, location_id, visibility,
	       allocation_method, status, closed_at, expires_at, created_at, updated_at
	FROM liquidations
`

// visibilityFilter is the shared predicate for every visibility-filtered read:
// a public liquidation, the viewer's own, or any restricted one when the viewer
// belongs to the mayorista roles the restriction targets. $1 is the viewer id
// and $2 the SeeRestricted flag.
const visibilityFilter = `(visibility = 'public' OR supplier_id = $1 OR ($2 AND visibility = 'private'))`

func scanLiquidations(rows pgx.Rows) ([]domain.Liquidation, error) {
	defer rows.Close()

	var liquidations []domain.Liquidation
	for rows.Next() {
		var liq domain.Liquidation
		if err := rows.Scan(
			&liq.ID, &liq.SupplierID, &liq.ProductName, &liq.Quantity, &liq.UnitOfMeasure,
			&liq.TotalPrice, &liq.UnitPrice, &liq.DeliveryTime, &liq.LocationID, &liq.Visibility,
			&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt,
			&liq.CreatedAt, &liq.UpdatedAt,
		); err != nil {
			return nil, err
		}
		liquidations = append(liquidations, liq)
	}

	return liquidations, rows.Err()
}

func (r *LiquidationRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Liquidation, error) {
	query := liquidationColumns + ` WHERE id = $1`

	var liq domain.Liquidation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&liq.ID, &liq.SupplierID, &liq.ProductName, &liq.Quantity, &liq.UnitOfMeasure,
		&liq.TotalPrice, &liq.UnitPrice, &liq.DeliveryTime, &liq.LocationID, &liq.Visibility,
		&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt,
		&liq.CreatedAt, &liq.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("liquidation.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("liquidation.FindByID: %w", err)
	}

	return &liq, nil
}

func (r *LiquidationRepositoryImpl) FindVisibleByID(ctx context.Context, id uuid.UUID, viewer port.LiquidationViewer) (*domain.Liquidation, error) {
	query := liquidationColumns + ` WHERE id = $3 AND ` + visibilityFilter

	var liq domain.Liquidation
	err := r.pool.QueryRow(ctx, query, viewer.ID, viewer.SeeRestricted, id).Scan(
		&liq.ID, &liq.SupplierID, &liq.ProductName, &liq.Quantity, &liq.UnitOfMeasure,
		&liq.TotalPrice, &liq.UnitPrice, &liq.DeliveryTime, &liq.LocationID, &liq.Visibility,
		&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt,
		&liq.CreatedAt, &liq.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("liquidation.FindVisibleByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("liquidation.FindVisibleByID: %w", err)
	}

	return &liq, nil
}

func (r *LiquidationRepositoryImpl) FindBySupplier(ctx context.Context, supplierID uuid.UUID, viewer port.LiquidationViewer) ([]domain.Liquidation, error) {
	query := liquidationColumns + ` WHERE supplier_id = $3 AND ` + visibilityFilter + ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, viewer.ID, viewer.SeeRestricted, supplierID)
	if err != nil {
		return nil, fmt.Errorf("liquidation.FindBySupplier: %w", err)
	}

	liquidations, err := scanLiquidations(rows)
	if err != nil {
		return nil, fmt.Errorf("liquidation.FindBySupplier: %w", err)
	}

	return liquidations, nil
}

func (r *LiquidationRepositoryImpl) FindOpen(ctx context.Context, viewer port.LiquidationViewer) ([]domain.Liquidation, error) {
	query := liquidationColumns + ` WHERE status = 'open' AND ` + visibilityFilter + ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, viewer.ID, viewer.SeeRestricted)
	if err != nil {
		return nil, fmt.Errorf("liquidation.FindOpen: %w", err)
	}

	liquidations, err := scanLiquidations(rows)
	if err != nil {
		return nil, fmt.Errorf("liquidation.FindOpen: %w", err)
	}

	return liquidations, nil
}

func (r *LiquidationRepositoryImpl) Save(ctx context.Context, liq *domain.Liquidation) error {
	query := `
		INSERT INTO liquidations (id, supplier_id, product_name, quantity, unit_of_measure, 
		                          total_price, unit_price, delivery_time, location_id, visibility,
		                          allocation_method, status, closed_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.pool.Exec(ctx, query,
		liq.ID, liq.SupplierID, liq.ProductName, liq.Quantity, liq.UnitOfMeasure,
		liq.TotalPrice, liq.UnitPrice, liq.DeliveryTime, liq.LocationID, liq.Visibility,
		liq.AllocationMethod, liq.Status, liq.ClosedAt, liq.ExpiresAt,
		liq.CreatedAt, liq.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("liquidation.Save: %w", err)
	}
	return nil
}

func (r *LiquidationRepositoryImpl) Update(ctx context.Context, liq *domain.Liquidation) error {
	query := `
		UPDATE liquidations
		SET product_name = $1, quantity = $2, unit_of_measure = $3, 
		    total_price = $4, unit_price = $5, delivery_time = $6, 
		    location_id = $7, visibility = $8, allocation_method = $9,
		    status = $10, closed_at = $11, expires_at = $12, updated_at = $13
		WHERE id = $14
	`
	_, err := r.pool.Exec(ctx, query,
		liq.ProductName, liq.Quantity, liq.UnitOfMeasure,
		liq.TotalPrice, liq.UnitPrice, liq.DeliveryTime,
		liq.LocationID, liq.Visibility, liq.AllocationMethod,
		liq.Status, liq.ClosedAt, liq.ExpiresAt, liq.UpdatedAt,
		liq.ID,
	)
	if err != nil {
		return fmt.Errorf("liquidation.Update: %w", err)
	}
	return nil
}

func (r *LiquidationRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM liquidations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("liquidation.Delete: %w", err)
	}
	return nil
}

var _ port.LiquidationRepository = (*LiquidationRepositoryImpl)(nil)
