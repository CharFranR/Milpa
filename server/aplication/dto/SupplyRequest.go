package dto

import (
	domain "milpa/domain/entities"
	"time"

	"github.com/google/uuid"
)

type SupplyRequestDTO struct {
	ID                   *uuid.UUID                 `json:"id"`
	BuyerID              *uuid.UUID                 `json:"buyer_id"`
	ProductName          string                     `json:"product_name"`
	TotalAmount          float64                    `json:"total_amount"`
	ActualAmount         float64                    `json:"actual_amount"`
	AmountUnit           domain.MeasurementOptions  `json:"amount_unit"`
	NumberOfUnits        float64                    `json:"number_of_units"`
	AmountPerUnit        float64                    `json:"amount_per_unit"`
	UnitOfMeasure        domain.MeasurementOptions  `json:"unit_of_measure"`
	Address              domain.Address             `json:"address"`
	RequestDeadline      time.Time                  `json:"request_deadline"`
	DeliveryDeadline     time.Time                  `json:"delivery_deadline"`
	Description          string                     `json:"description"`
	MultipleProviders    bool                       `json:"multiple_providers"`
	MinAmountPerProvider float64                    `json:"min_amount_per_provider"`
	Status               domain.SupplyRequestStatus `json:"status"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
}

type SupplyUpdateAmountsDTO struct {
	ID                   *uuid.UUID                `json:"id"`
	TotalAmount          float64                   `json:"total_amount"`
	ActualAmount         float64                   `json:"actual_amount"`
	AmountUnit           domain.MeasurementOptions `json:"amount_unit"`
	AmountPerUnit        float64                   `json:"amount_per_unit"`
	UnitOfMeasure        domain.MeasurementOptions `json:"unit_of_measure"`
	MultipleProviders    bool                      `json:"multiple_providers"`
	MinAmountPerProvider float64                   `json:"min_amount_per_provider"`
}

type SupplyUpdateTimeDTO struct {
	ID               *uuid.UUID `json:"id"`
	RequestDeadline  time.Time  `json:"request_deadline"`
	DeliveryDeadline time.Time  `json:"delivery_deadline"`
}

type SupplyGeneralUpdateDTO struct {
	ID                   *uuid.UUID                `json:"id"`
	ProductName          string                    `json:"product_name"`
	TotalAmount          float64                   `json:"total_amount"`
	ActualAmount         float64                   `json:"actual_amount"`
	AmountUnit           domain.MeasurementOptions `json:"amount_unit"`
	NumberOfUnits        float64                   `json:"number_of_units"`
	AmountPerUnit        float64                   `json:"amount_per_unit"`
	UnitOfMeasure        domain.MeasurementOptions `json:"unit_of_measure"`
	Address              domain.Address            `json:"address"`
	RequestDeadline      time.Time                 `json:"request_deadline"`
	DeliveryDeadline     time.Time                 `json:"delivery_deadline"`
	Description          string                    `json:"description"`
	MultipleProviders    bool                      `json:"multiple_providers"`
	MinAmountPerProvider float64                   `json:"min_amount_per_provider"`
}
