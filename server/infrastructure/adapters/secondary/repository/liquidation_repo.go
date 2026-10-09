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
	       allocation_method, status, closed_at, expires_at, assigned_buyer_id, created_at, updated_at
	FROM liquidations
`

const visibilityFilter = `(visibility = 'public' OR supplier_id = $1 OR ($2 AND visibility = 'wholesale') OR ($3 AND visibility = 'wholesale_retail') OR ($4 AND visibility = 'wholesale_corporate'))`

func liquidationViewerArgs(viewer port.LiquidationViewer) []any {
	return []any{viewer.ID, viewer.SeesWholesale(), viewer.SeesWholesaleRetail(), viewer.SeesWholesaleCorporate()}
}

func scanLiquidations(rows pgx.Rows) ([]domain.Liquidation, error) {
	defer rows.Close()

	var liquidations []domain.Liquidation
	for rows.Next() {
		var liq domain.Liquidation
		if err := rows.Scan(
			&liq.ID, &liq.SupplierID, &liq.ProductName, &liq.Quantity, &liq.UnitOfMeasure,
			&liq.TotalPrice, &liq.UnitPrice, &liq.DeliveryTime, &liq.LocationID, &liq.Visibility,
			&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt, &liq.AssignedBuyerID,
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
		&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt, &liq.AssignedBuyerID,
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
	query := liquidationColumns + ` WHERE id = $5 AND ` + visibilityFilter

	args := append(liquidationViewerArgs(viewer), id)

	var liq domain.Liquidation
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&liq.ID, &liq.SupplierID, &liq.ProductName, &liq.Quantity, &liq.UnitOfMeasure,
		&liq.TotalPrice, &liq.UnitPrice, &liq.DeliveryTime, &liq.LocationID, &liq.Visibility,
		&liq.AllocationMethod, &liq.Status, &liq.ClosedAt, &liq.ExpiresAt, &liq.AssignedBuyerID,
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
	query := liquidationColumns + ` WHERE supplier_id = $5 AND ` + visibilityFilter + ` ORDER BY created_at DESC`

	args := append(liquidationViewerArgs(viewer), supplierID)

	rows, err := r.pool.Query(ctx, query, args...)
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

	rows, err := r.pool.Query(ctx, query, liquidationViewerArgs(viewer)...)
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
		                          allocation_method, status, closed_at, expires_at, assigned_buyer_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.pool.Exec(ctx, query,
		liq.ID, liq.SupplierID, liq.ProductName, liq.Quantity, liq.UnitOfMeasure,
		liq.TotalPrice, liq.UnitPrice, liq.DeliveryTime, liq.LocationID, liq.Visibility,
		liq.AllocationMethod, liq.Status, liq.ClosedAt, liq.ExpiresAt, liq.AssignedBuyerID,
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
		    status = $10, closed_at = $11, expires_at = $12, assigned_buyer_id = $13, updated_at = $14
		WHERE id = $15
	`
	_, err := r.pool.Exec(ctx, query,
		liq.ProductName, liq.Quantity, liq.UnitOfMeasure,
		liq.TotalPrice, liq.UnitPrice, liq.DeliveryTime,
		liq.LocationID, liq.Visibility, liq.AllocationMethod,
		liq.Status, liq.ClosedAt, liq.ExpiresAt, liq.AssignedBuyerID, liq.UpdatedAt,
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

func (r *LiquidationRepositoryImpl) SaveInterest(ctx context.Context, interest *domain.LiquidationInterest) error {
	query := `
		INSERT INTO liquidation_interests (id, liquidation_id, buyer_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, interest.ID, interest.LiquidationID, interest.BuyerID, interest.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("liquidation.SaveInterest: %w", domain.ErrInterestAlreadyExists)
		}
		return fmt.Errorf("liquidation.SaveInterest: %w", err)
	}
	return nil
}

func (r *LiquidationRepositoryImpl) FindInterests(ctx context.Context, liquidationID uuid.UUID) ([]domain.LiquidationInterest, error) {
	query := `
		SELECT id, liquidation_id, buyer_id, created_at
		FROM liquidation_interests
		WHERE liquidation_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, liquidationID)
	if err != nil {
		return nil, fmt.Errorf("liquidation.FindInterests: %w", err)
	}
	defer rows.Close()

	var interests []domain.LiquidationInterest
	for rows.Next() {
		var interest domain.LiquidationInterest
		if err := rows.Scan(&interest.ID, &interest.LiquidationID, &interest.BuyerID, &interest.CreatedAt); err != nil {
			return nil, fmt.Errorf("liquidation.FindInterests: %w", err)
		}
		interests = append(interests, interest)
	}

	return interests, rows.Err()
}

func (r *LiquidationRepositoryImpl) InterestExists(ctx context.Context, liquidationID, buyerID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM liquidation_interests WHERE liquidation_id = $1 AND buyer_id = $2)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, liquidationID, buyerID).Scan(&exists); err != nil {
		return false, fmt.Errorf("liquidation.InterestExists: %w", err)
	}

	return exists, nil
}

var _ port.LiquidationRepository = (*LiquidationRepositoryImpl)(nil)
