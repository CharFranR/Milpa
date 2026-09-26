package dto

import (
	domain "milpa/domain/entities"
	"time"

	"github.com/google/uuid"
)

type SupplyRequestDTO struct {
	ID                   *uuid.UUID                `json:"id"`
	BuyerID              *uuid.UUID                `json:"buyer_id"`
	ProductName          string                    `json:"product_name"`
	TotalAmount          float32                   `json:"total_amount"`
	ActualAomunt         float32                   `json:"actual_amount"`
	AmountUnit           domain.MeasurementOptions `json:"amount_measure"`
	NumberOfUnits        float32                   `json:"numer_units"`
	AmountPerUnit        float32                   `json:"amount_unit"`
	UnitOfMeasure        domain.MeasurementOptions `json:"unit_measure"`
	Address              domain.Address            `json:"Addrres"`
	RequestDeadline      time.Time                 `json:"request_deadline"`
	DeliveryDeadline     time.Time                 `json:"delivery_deadline"`
	Description          string                    `json:"description"`
	MultipleProvidres    bool                      `json:"multiple_providers"`
	MinAmountPerProvider float64                   `json:"min_amount_provider"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

type SupplyUpdateAmountsDTO struct {
	ID                   *uuid.UUID                `json:"id"`
	TotalAmount          float32                   `json:"total_amount"`
	ActualAomunt         float32                   `json:"actual_amount"`
	AmountUnit           domain.MeasurementOptions `json:"amount_measure"`
	AmountPerUnit        float32                   `json:"amount_unit"`
	UnitOfMeasure        domain.MeasurementOptions `json:"unit_measure"`
	MultipleProvidres    bool                      `json:"multiple_providers"`
	MinAmountPerProvider float64                   `json:"min_amount_provider"`
}

type SupplyUpdateTimeDTO struct {
	ID               *uuid.UUID `json:"id"`
	RequestDeadline  time.Time  `json:"request_deadline"`
	DeliveryDeadline time.Time  `json:"delivery_deadline"`
}

type SupplyGeneralUpdateDTO struct {
	ID                   *uuid.UUID                `json:"id"`
	ProductName          string                    `json:"product_name"`
	TotalAmount          float32                   `json:"total_amount"`
	AmountUnit           float32                   `json:"actual_amount"`
	AmountUOF            domain.MeasurementOptions `json:"amount_measure"`
	NumberOfUnits        float32                   `json:"numer_units"`
	AmountPerUnit        float32                   `json:"amount_unit"`
	UnitOfMeasure        domain.MeasurementOptions `json:"unit_measure"`
	Address              domain.Address            `json:"Addrres"`
	RequestDeadline      time.Time                 `json:"request_deadline"`
	DeliveryDeadline     time.Time                 `json:"delivery_deadline"`
	Description          string                    `json:"description"`
	MultipleProvidres    bool                      `json:"multiple_providers"`
	MinAmountPerProvider float64                   `json:"min_amount_provider"`
}
