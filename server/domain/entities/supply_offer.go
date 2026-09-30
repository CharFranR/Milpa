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

// ValidMeasurementOptions reports whether m is part of the unit vocabulary.
//
// It lives next to the constants so the vocabulary has exactly one definition,
// the same way AllocationMethod.String and AllocationMethod.Scan do it in
// liquidation.go. A MeasurementOptions is an iota and the field arrives straight
// off the wire, so any integer the client sends is representable — including
// ones nobody defined. Without this there is no way to tell a kilogram from a
// request for unit 7.
func ValidMeasurementOptions(m MeasurementOptions) bool {
	switch m {
	case Kg, Lb, Tn:
		return true
	default:
		return false
	}
}

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
	TotalAmount   float64
	AmountUnit    MeasurementOptions

	ProposedDeliveryDay time.Time
	DeliveryAvailable   bool
	Status              OfferStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSupplyOffer(
	SupplierID uuid.UUID, SupplyRequest uuid.UUID, TotalAmount float64, AmountUnit MeasurementOptions, ProposedDeliveryDay time.Time, DeliveryAvailable bool,
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
