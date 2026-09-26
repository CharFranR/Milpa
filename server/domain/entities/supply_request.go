package domain

import (
	"time"

	"github.com/google/uuid"
)

type SupplyRequest struct {
	ID                   uuid.UUID
	BuyerID              uuid.UUID
	ProductName          string
	TotalAmount          float32
	ActualAomunt         float32
	AmountUnit           MeasurementOptions
	NumberOfUnits        float32
	AmountPerUnit        float32
	UnitOfMeasure        MeasurementOptions
	Address              Address
	RequestDeadline      time.Time
	DeliveryDeadline     time.Time
	Description          string
	MultipleProvidres    bool
	MinAmountPerProvider float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSupplyRequest(
	BuyerID uuid.UUID, ProductName string, TotalAmount float32, AmountUnit MeasurementOptions, NumberOfUnits float32,
	AmountPerUnit float32, UnitOfMeasure MeasurementOptions,
	Address Address, RequestDeadline time.Time, DeliveryDeadline time.Time, Description string, MultipleProvidres bool,
) *SupplyRequest {

	return &SupplyRequest{
		ID:                uuid.New(),
		BuyerID:           BuyerID,
		ProductName:       ProductName,
		TotalAmount:       TotalAmount,
		ActualAomunt:      TotalAmount,
		AmountUnit:        AmountUnit,
		NumberOfUnits:     NumberOfUnits,
		AmountPerUnit:     AmountPerUnit,
		UnitOfMeasure:     UnitOfMeasure,
		Address:           Address,
		RequestDeadline:   RequestDeadline,
		DeliveryDeadline:  DeliveryDeadline,
		Description:       Description,
		MultipleProvidres: MultipleProvidres,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

}
