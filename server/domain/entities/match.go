package domain

import (
	"time"

	"github.com/google/uuid"
)

type MatchStatus int

const (
	MatchActive MatchStatus = iota
	MatchCancelled
)

type Match struct {
	ID            uuid.UUID
	SupplyOffer   uuid.UUID
	SupplyRequest uuid.UUID

	Status MatchStatus

	MatchedAmount float64
	AmountUnit    MeasurementOptions

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMatch(supplyOffer uuid.UUID, supplyRequest uuid.UUID, matchedAmount float64, unit MeasurementOptions) *Match {
	return &Match{
		ID:            uuid.New(),
		SupplyOffer:   supplyOffer,
		SupplyRequest: supplyRequest,
		Status:        MatchActive,
		MatchedAmount: matchedAmount,
		AmountUnit:    unit,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (m Match) IsActive() bool {
	return m.Status == MatchActive
}

func (m *Match) Cancel() error {
	if m.Status != MatchActive {
		return ErrInvalidMatchStatus
	}
	m.Status = MatchCancelled
	m.UpdatedAt = time.Now()
	return nil
}

func (m MatchStatus) String() string {
	switch m {
	case MatchActive:
		return "active"
	case MatchCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}
