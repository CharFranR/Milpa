package domain

import (
	"time"

	"github.com/google/uuid"
)

type SupplierInventory struct {
	ID          uuid.UUID
	SupplierID  uuid.UUID
	ProductName string
	Quantity    float64
	AmountUnit  MeasurementOptions

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSupplierInventory(supplierID uuid.UUID, productName string, quantity float64, unit MeasurementOptions) *SupplierInventory {
	return &SupplierInventory{
		ID:          uuid.New(),
		SupplierID:  supplierID,
		ProductName: productName,
		Quantity:    quantity,
		AmountUnit:  unit,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (s *SupplierInventory) SetQuantity(quantity float64) error {
	if quantity < 0 {
		return ErrInvalidQuantity
	}
	s.Quantity = quantity
	s.UpdatedAt = time.Now()
	return nil
}
