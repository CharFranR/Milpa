package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LiquidationStatus int

const (
	LiquidationOpen LiquidationStatus = iota
	LiquidationClosed
	LiquidationExpired
	LiquidationAssigned
)

// AllocationMethod defines how a liquidation is assigned.
type AllocationMethod int

const (
	AllocationManual AllocationMethod = iota
)

type Liquidation struct {
	ID               uuid.UUID
	SupplierID       uuid.UUID
	ProductName      string
	Quantity         float64
	UnitOfMeasure    string
	TotalPrice       float64
	UnitPrice        float64
	DeliveryTime     string
	LocationID       uuid.UUID
	Visibility       string
	AllocationMethod AllocationMethod
	Status           LiquidationStatus
	ClosedAt         *time.Time
	ExpiresAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Builder

func NewLiquidation(supplierID uuid.UUID, productName string, quantity float64, unitOfMeasure string, totalPrice, unitPrice float64, now time.Time) (*Liquidation, error) {
	if productName == "" {
		return nil, ErrProductNameRequired
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if unitOfMeasure == "" {
		return nil, ErrUnitOfMeasureRequired
	}
	if totalPrice <= 0 {
		return nil, ErrInvalidPrice
	}
	if unitPrice <= 0 {
		return nil, ErrInvalidPrice
	}

	return &Liquidation{
		ID:               uuid.New(),
		SupplierID:       supplierID,
		ProductName:      productName,
		Quantity:         quantity,
		UnitOfMeasure:    unitOfMeasure,
		TotalPrice:       totalPrice,
		UnitPrice:        unitPrice,
		Visibility:       "public",
		AllocationMethod: AllocationManual,
		Status:           LiquidationOpen,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// Get

func (l Liquidation) IsOpen() bool {
	return l.Status == LiquidationOpen
}

func (l Liquidation) CanReceiveInterest() bool {
	return l.Status == LiquidationOpen
}

func (l Liquidation) IsExpired(now time.Time) bool {
	if l.ExpiresAt == nil {
		return false
	}
	return now.After(*l.ExpiresAt)
}

// Set

func (l *Liquidation) Close(now time.Time) error {
	if l.Status != LiquidationOpen {
		return ErrLiquidationNotOpen
	}
	l.Status = LiquidationClosed
	l.ClosedAt = &now
	l.Touch(now)
	return nil
}

func (l *Liquidation) Expire(now time.Time) error {
	if l.Status != LiquidationOpen {
		return ErrLiquidationNotOpen
	}
	l.Status = LiquidationExpired
	l.ClosedAt = &now
	l.Touch(now)
	return nil
}

func (l *Liquidation) Assign(now time.Time) error {
	if l.Status != LiquidationOpen {
		return ErrLiquidationCannotAssign
	}
	l.Status = LiquidationAssigned
	l.ClosedAt = &now
	l.Touch(now)
	return nil
}

func (l *Liquidation) Touch(now time.Time) {
	l.UpdatedAt = now
}

func (l *Liquidation) UpdateDeliveryTime(deliveryTime string, now time.Time) {
	l.DeliveryTime = deliveryTime
	l.Touch(now)
}

func (l *Liquidation) UpdateVisibility(visibility string, now time.Time) error {
	if visibility != "public" && visibility != "private" {
		return ErrInvalidVisibility
	}
	l.Visibility = visibility
	l.Touch(now)
	return nil
}

func (l *Liquidation) SetExpiry(expiresAt time.Time, now time.Time) {
	l.ExpiresAt = &expiresAt
	l.Touch(now)
}

// String

func (s LiquidationStatus) String() string {
	switch s {
	case LiquidationOpen:
		return "open"
	case LiquidationClosed:
		return "closed"
	case LiquidationExpired:
		return "expired"
	case LiquidationAssigned:
		return "assigned"
	default:
		return "unknown"
	}
}

func (a AllocationMethod) String() string {
	switch a {
	case AllocationManual:
		return "manual"
	default:
		return "unknown"
	}
}

func (s *LiquidationStatus) Scan(src any) error {
	text, err := scanText(src)
	if err != nil {
		return fmt.Errorf("liquidation status: %w", err)
	}

	for _, candidate := range []LiquidationStatus{LiquidationOpen, LiquidationClosed, LiquidationExpired, LiquidationAssigned} {
		if candidate.String() == text {
			*s = candidate
			return nil
		}
	}

	return fmt.Errorf("liquidation status: unknown value %q", text)
}

func (a *AllocationMethod) Scan(src any) error {
	text, err := scanText(src)
	if err != nil {
		return fmt.Errorf("liquidation allocation method: %w", err)
	}

	if text == AllocationManual.String() {
		*a = AllocationManual
		return nil
	}

	return fmt.Errorf("liquidation allocation method: unknown value %q", text)
}

// scanText normalises what a driver may hand to Scan.
func scanText(src any) (string, error) {
	switch v := src.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("cannot scan %T", src)
	}
}
