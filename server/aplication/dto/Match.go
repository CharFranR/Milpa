package dto

import (
	domain "milpa/domain/entities"
	"time"

	"github.com/google/uuid"
)

type MatchDTO struct {
	ID            uuid.UUID                 `json:"id"`
	SupplyOffer   uuid.UUID                 `json:"supply_offer"`
	SupplyRequest uuid.UUID                 `json:"supply_request"`
	Status        domain.MatchStatusOptions `json:"status"`
	MatchedAmount float32                   `json:"matched_amount"`
	AmountUnit    domain.MeasurementOptions `json:"amount_unit"`
	Created_at    time.Time                 `json:"created_at"`
	Updated_at    time.Time                 `json:"updated_at"`
}

type MatchUpdateDTO struct {
	ID            uuid.UUID                 `json:"id"`
	Status        domain.MatchStatusOptions `json:"status"`
	MatchedAmount float32                   `json:"matched_amount"`
	AmountUnit    domain.MeasurementOptions `json:"amount_unit"`
}
