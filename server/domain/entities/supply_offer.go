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

type OfferStatus int

const (
	OfferActive OfferStatus = iota
	OfferMatched
	OfferRejected
	OfferWithdrawn
)

type SupplyOffer struct {
	ID            uuid.UUID
	SupplierID    uuid.UUID
	SupplyRequest uuid.UUID
	TotalAmount   float32
	AmountUnit    MeasurementOptions

	ProposedDeliveryDay time.Time
	DeliveryAvailable   bool
	Status              OfferStatus

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
		Status:              OfferActive,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

}

func (s SupplyOffer) IsActionable() bool {
	return s.Status == OfferActive
}

func (s *SupplyOffer) Withdraw() error {
	if s.Status != OfferActive {
		return ErrInvalidOfferStatus
	}
	s.Status = OfferWithdrawn
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyOffer) Reject() error {
	if s.Status != OfferActive {
		return ErrInvalidOfferStatus
	}
	s.Status = OfferRejected
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyOffer) MarkMatched() error {
	if s.Status != OfferActive {
		return ErrInvalidOfferStatus
	}
	s.Status = OfferMatched
	s.UpdatedAt = time.Now()
	return nil
}

func (s OfferStatus) String() string {
	switch s {
	case OfferActive:
		return "active"
	case OfferMatched:
		return "matched"
	case OfferRejected:
		return "rejected"
	case OfferWithdrawn:
		return "withdrawn"
	default:
		return "unknown"
	}
}
