package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type LiquidationDTO struct {
	ID               uuid.UUID                `json:"id"`
	SupplierID       uuid.UUID                `json:"supplier_id"`
	ProductName      string                   `json:"product_name"`
	Quantity         float64                  `json:"quantity"`
	UnitOfMeasure    string                   `json:"unit_of_measure"`
	TotalPrice       float64                  `json:"total_price"`
	UnitPrice        float64                  `json:"unit_price"`
	DeliveryTime     string                   `json:"delivery_time"`
	LocationID       uuid.UUID                `json:"location_id"`
	Visibility       string                   `json:"visibility"`
	AllocationMethod domain.AllocationMethod  `json:"allocation_method"`
	Status           domain.LiquidationStatus `json:"status"`
	ClosedAt         *time.Time               `json:"closed_at,omitempty"`
	ExpiresAt        *time.Time               `json:"expires_at,omitempty"`
	AssignedBuyerID  *uuid.UUID               `json:"assigned_buyer_id,omitempty"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type LiquidationInterestDTO struct {
	ID            uuid.UUID `json:"id"`
	LiquidationID uuid.UUID `json:"liquidation_id"`
	BuyerID       uuid.UUID `json:"buyer_id"`
	BuyerName     string    `json:"buyer_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateLiquidationRequest struct {
	ProductName      string                  `json:"product_name"`
	Quantity         float64                 `json:"quantity"`
	UnitOfMeasure    string                  `json:"unit_of_measure"`
	TotalPrice       float64                 `json:"total_price"`
	UnitPrice        float64                 `json:"unit_price"`
	DeliveryTime     string                  `json:"delivery_time,omitempty"`
	LocationID       uuid.UUID               `json:"location_id"`
	Visibility       string                  `json:"visibility,omitempty"`
	AllocationMethod domain.AllocationMethod `json:"allocation_method,omitempty"`
	ExpiresAt        *time.Time              `json:"expires_at,omitempty"`
}

type UpdateLiquidationRequest struct {
	ProductName      *string                  `json:"product_name,omitempty"`
	Quantity         *float64                 `json:"quantity,omitempty"`
	UnitOfMeasure    *string                  `json:"unit_of_measure,omitempty"`
	TotalPrice       *float64                 `json:"total_price,omitempty"`
	UnitPrice        *float64                 `json:"unit_price,omitempty"`
	DeliveryTime     *string                  `json:"delivery_time,omitempty"`
	LocationID       *uuid.UUID               `json:"location_id,omitempty"`
	Visibility       *string                  `json:"visibility,omitempty"`
	AllocationMethod *domain.AllocationMethod `json:"allocation_method,omitempty"`
	ExpiresAt        *time.Time               `json:"expires_at,omitempty"`
}

type AssignLiquidationRequest struct {
	BuyerID *uuid.UUID `json:"buyer_id,omitempty"`
}
