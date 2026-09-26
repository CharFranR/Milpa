package domain

import (
	"time"

	"github.com/google/uuid"
)

type MeasurementOptions int

const (
	Kg MeasurementOptions = iota
	Lb
	Tn
)

type SupplyOffer struct {
	ID            uuid.UUID
	SupplierID    uuid.UUID
	SupplyRequest uuid.UUID
	TotalAmount   float32
	AmountUnit    MeasurementOptions

	ProposedDeliveryDay time.Time
	DeliveryAvailable   bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSupplyOffer(
	SupplierID uuid.UUID, SupplyRequest uuid.UUID, TotalAmount float32, AmountUnit MeasurementOptions, ProposedDeliveryDay time.Time, DeliveryAvailable bool,
) *SupplyOffer {
	return &SupplyOffer{
		ID:                  uuid.New(),
		SupplierID:          SupplierID,
		SupplyRequest:       SupplyRequest,
		TotalAmount:         TotalAmount,
		AmountUnit:          AmountUnit,
		ProposedDeliveryDay: ProposedDeliveryDay,
		DeliveryAvailable:   DeliveryAvailable,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

}
