package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type SupplierInventoryDTO struct {
	ID          *uuid.UUID                `json:"id"`
	SupplierID  *uuid.UUID                `json:"supplier_id"`
	ProductName string                    `json:"product_name"`
	Quantity    float64                   `json:"quantity"`
	AmountUnit  domain.MeasurementOptions `json:"measurement"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type UpsertSupplierInventoryRequest struct {
	ProductName string                    `json:"product_name"`
	Quantity    float64                   `json:"quantity"`
	AmountUnit  domain.MeasurementOptions `json:"measurement"`
}
