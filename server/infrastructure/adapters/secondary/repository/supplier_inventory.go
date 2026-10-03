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

type SupplierInventoryRepositoryImpl struct {
	db Querier
}

func NewSupplierInventoryRepository(pool DB) *SupplierInventoryRepositoryImpl {
	return &SupplierInventoryRepositoryImpl{db: pool}
}

func scanSupplierInventory(scan func(dest ...any) error) (domain.SupplierInventory, error) {
	var inventory domain.SupplierInventory

	err := scan(
		&inventory.ID, &inventory.SupplierID, &inventory.ProductName, &inventory.Quantity, &inventory.AmountUnit,
		&inventory.CreatedAt, &inventory.UpdatedAt,
	)
	if err != nil {
		return domain.SupplierInventory{}, err
	}

	return inventory, nil
}

func (r *SupplierInventoryRepositoryImpl) Create(ctx context.Context, inventory *domain.SupplierInventory) error {
	query := `
		INSERT INTO supplier_inventory (id, supplier_id, product_name, quantity, amount_unit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		inventory.ID, inventory.SupplierID, inventory.ProductName, inventory.Quantity, inventory.AmountUnit,
		inventory.CreatedAt, inventory.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("supplierInventory.Create: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("supplierInventory.Create: %w", err)
	}
	return nil
}

func (r *SupplierInventoryRepositoryImpl) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplierInventory, error) {
	query := `
		SELECT id, supplier_id, product_name, quantity, amount_unit, created_at, updated_at
		FROM supplier_inventory
		WHERE supplier_id = $1
	`

	rows, err := r.db.Query(ctx, query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("supplierInventory.ListBySupplier: %w", err)
	}
	defer rows.Close()

	var inventories []domain.SupplierInventory
	for rows.Next() {
		inventory, err := scanSupplierInventory(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("supplierInventory.ListBySupplier: %w", err)
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("supplierInventory.ListBySupplier: %w", err)
	}

	return inventories, nil
}

func (r *SupplierInventoryRepositoryImpl) FindBySupplierAndProduct(ctx context.Context, supplierID uuid.UUID, productName string) (domain.SupplierInventory, error) {
	query := `
		SELECT id, supplier_id, product_name, quantity, amount_unit, created_at, updated_at
		FROM supplier_inventory
		WHERE supplier_id = $1 AND product_name = $2
	`

	inventory, err := scanSupplierInventory(r.db.QueryRow(ctx, query, supplierID, productName).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplierInventory{}, fmt.Errorf("supplierInventory.FindBySupplierAndProduct: %w", domain.ErrNotFound)
		}
		return domain.SupplierInventory{}, fmt.Errorf("supplierInventory.FindBySupplierAndProduct: %w", err)
	}

	return inventory, nil
}

func (r *SupplierInventoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (domain.SupplierInventory, error) {
	query := `
		SELECT id, supplier_id, product_name, quantity, amount_unit, created_at, updated_at
		FROM supplier_inventory
		WHERE id = $1
	`

	inventory, err := scanSupplierInventory(r.db.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SupplierInventory{}, fmt.Errorf("supplierInventory.GetByID: %w", domain.ErrNotFound)
		}
		return domain.SupplierInventory{}, fmt.Errorf("supplierInventory.GetByID: %w", err)
	}

	return inventory, nil
}

func (r *SupplierInventoryRepositoryImpl) Update(ctx context.Context, inventory *domain.SupplierInventory) error {
	query := `
		UPDATE supplier_inventory
		SET product_name = $1, quantity = $2, amount_unit = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query,
		inventory.ProductName, inventory.Quantity, inventory.AmountUnit, inventory.UpdatedAt, inventory.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("supplierInventory.Update: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("supplierInventory.Update: %w", err)
	}
	return nil
}

func (r *SupplierInventoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM supplier_inventory WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("supplierInventory.Delete: %w", err)
	}
	return nil
}

var _ port.SupplierInventoryRepository = (*SupplierInventoryRepositoryImpl)(nil)
