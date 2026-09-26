package domain

import (
	"time"

	"github.com/google/uuid"
)

type MatchStatusOptions int

const (
	Pending MatchStatusOptions = iota
	Active
	Cancelled
)

type Match struct {
	ID            uuid.UUID
	SupplyOffer   uuid.UUID
	SupplyRequest uuid.UUID

	Status MatchStatusOptions

	MatchedAmount float32
	AmountUnit    MeasurementOptions

	Created_at time.Time
	Updated_at time.Time
}

func NewMatch(id uuid.UUID, supplyOffer uuid.UUID, supplyRequest uuid.UUID, status MatchStatusOptions) *Match {
	return &Match{
		ID:            id,
		SupplyOffer:   supplyOffer,
		SupplyRequest: supplyRequest,
		Status:        status,
		Created_at:    time.Now(),
		Updated_at:    time.Now(),
	}
}
