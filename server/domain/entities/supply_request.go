package domain

import (
	"time"

	"github.com/google/uuid"
)

type SupplyRequestStatus int

const (
	SupplyRequestOpen SupplyRequestStatus = iota
	SupplyRequestCancelled
	SupplyRequestCompleted
	SupplyRequestExpired
)

type SupplyRequest struct {
	ID                   uuid.UUID
	BuyerID              uuid.UUID
	ProductName          string
	TotalAmount          float32
	ActualAmount         float32
	AmountUnit           MeasurementOptions
	NumberOfUnits        float32
	AmountPerUnit        float32
	UnitOfMeasure        MeasurementOptions
	Address              Address
	RequestDeadline      time.Time
	DeliveryDeadline     time.Time
	Description          string
	MultipleProviders    bool
	MinAmountPerProvider float64
	Status               SupplyRequestStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSupplyRequest(
	BuyerID uuid.UUID, ProductName string, TotalAmount float32, AmountUnit MeasurementOptions, NumberOfUnits float32,
	AmountPerUnit float32, UnitOfMeasure MeasurementOptions,
	Address Address, RequestDeadline time.Time, DeliveryDeadline time.Time, Description string, MultipleProviders bool,
) *SupplyRequest {

	return &SupplyRequest{
		ID:                uuid.New(),
		BuyerID:           BuyerID,
		ProductName:       ProductName,
		TotalAmount:       TotalAmount,
		ActualAmount:      TotalAmount,
		AmountUnit:        AmountUnit,
		NumberOfUnits:     NumberOfUnits,
		AmountPerUnit:     AmountPerUnit,
		UnitOfMeasure:     UnitOfMeasure,
		Address:           Address,
		RequestDeadline:   RequestDeadline,
		DeliveryDeadline:  DeliveryDeadline,
		Description:       Description,
		MultipleProviders: MultipleProviders,
		Status:            SupplyRequestOpen,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

}

func (s SupplyRequest) IsOpen() bool {
	return s.Status == SupplyRequestOpen
}

func (s SupplyRequest) RemainingAmount() float32 {
	return s.ActualAmount
}

func (s *SupplyRequest) Cancel() error {
	if s.Status != SupplyRequestOpen {
		return ErrInvalidRequestStatus
	}
	s.Status = SupplyRequestCancelled
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyRequest) Complete() error {
	if s.Status != SupplyRequestOpen {
		return ErrInvalidRequestStatus
	}
	s.Status = SupplyRequestCompleted
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyRequest) Expire() error {
	if s.Status != SupplyRequestOpen {
		return ErrInvalidRequestStatus
	}
	s.Status = SupplyRequestExpired
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyRequest) ReserveAmount(amount float32) error {
	if s.Status != SupplyRequestOpen {
		return ErrInvalidRequestStatus
	}
	if amount <= 0 {
		return ErrInvalidInput
	}
	if amount > s.ActualAmount {
		return ErrInsufficientAmount
	}
	s.ActualAmount -= amount
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SupplyRequest) ReleaseAmount(amount float32) error {
	if amount <= 0 {
		return ErrInvalidInput
	}
	s.ActualAmount += amount
	if s.ActualAmount > s.TotalAmount {
		s.ActualAmount = s.TotalAmount
	}
	if s.ActualAmount < 0 {
		s.ActualAmount = 0
	}
	s.UpdatedAt = time.Now()
	return nil
}

func (s SupplyRequestStatus) String() string {
	switch s {
	case SupplyRequestOpen:
		return "open"
	case SupplyRequestCancelled:
		return "cancelled"
	case SupplyRequestCompleted:
		return "completed"
	case SupplyRequestExpired:
		return "expired"
	default:
		return "unknown"
	}
}
