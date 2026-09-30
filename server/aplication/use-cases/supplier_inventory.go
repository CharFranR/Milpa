package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

type SupplierInventoryUseCaseImpl struct {
	inventoryRepo port.SupplierInventoryRepository
}

func NewSupplierInventoryUseCase(inventoryRepo port.SupplierInventoryRepository) *SupplierInventoryUseCaseImpl {
	return &SupplierInventoryUseCaseImpl{inventoryRepo: inventoryRepo}
}

func (uc *SupplierInventoryUseCaseImpl) Upsert(ctx context.Context, supplierID uuid.UUID, req dto.UpsertSupplierInventoryRequest) (*dto.SupplierInventoryDTO, error) {
	productName := strings.TrimSpace(req.ProductName)
	if productName == "" {
		return nil, fmt.Errorf("%w: product_name is required", domain.ErrInvalidInput)
	}
	if !domain.ValidMeasurementOptions(req.AmountUnit) {
		return nil, fmt.Errorf("%w: unknown unit of measure %d", domain.ErrInvalidInput, int(req.AmountUnit))
	}

	existing, err := uc.inventoryRepo.FindBySupplierAndProduct(ctx, supplierID, productName)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	if errors.Is(err, domain.ErrNotFound) {
		inventory := domain.NewSupplierInventory(supplierID, productName, 0, req.AmountUnit)
		if err := inventory.SetQuantity(req.Quantity); err != nil {
			return nil, err
		}
		if err := uc.inventoryRepo.Create(ctx, inventory); err != nil {
			return nil, err
		}
		return inventoryToDTO(inventory), nil
	}

	if err := existing.SetQuantity(req.Quantity); err != nil {
		return nil, err
	}
	existing.AmountUnit = req.AmountUnit
	if err := uc.inventoryRepo.Update(ctx, &existing); err != nil {
		return nil, err
	}
	return inventoryToDTO(&existing), nil
}

func (uc *SupplierInventoryUseCaseImpl) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplierInventoryDTO, error) {
	inventories, err := uc.inventoryRepo.ListBySupplier(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.SupplierInventoryDTO, 0, len(inventories))
	for i := range inventories {
		result = append(result, inventoryToDTO(&inventories[i]))
	}
	return result, nil
}

func (uc *SupplierInventoryUseCaseImpl) Delete(ctx context.Context, supplierID uuid.UUID, id uuid.UUID) error {
	inventory, err := uc.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if inventory.SupplierID != supplierID {
		return domain.ErrForbidden
	}
	return uc.inventoryRepo.Delete(ctx, id)
}

func inventoryToDTO(inventory *domain.SupplierInventory) *dto.SupplierInventoryDTO {
	return &dto.SupplierInventoryDTO{
		ID:          &inventory.ID,
		SupplierID:  &inventory.SupplierID,
		ProductName: inventory.ProductName,
		Quantity:    inventory.Quantity,
		AmountUnit:  inventory.AmountUnit,
		CreatedAt:   inventory.CreatedAt,
		UpdatedAt:   inventory.UpdatedAt,
	}
}

var _ primary.SupplierInventoryUseCase = (*SupplierInventoryUseCaseImpl)(nil)
