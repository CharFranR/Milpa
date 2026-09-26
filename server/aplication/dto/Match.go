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
	Status        domain.MatchStatus        `json:"status"`
	MatchedAmount float32                   `json:"matched_amount"`
	AmountUnit    domain.MeasurementOptions `json:"amount_unit"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

type MatchUpdateDTO struct {
	ID            uuid.UUID                 `json:"id"`
	Status        domain.MatchStatus        `json:"status"`
	MatchedAmount float32                   `json:"matched_amount"`
	AmountUnit    domain.MeasurementOptions `json:"amount_unit"`
}

type MatchCreatedDTO struct {
	Match       *MatchDTO       `json:"match"`
	Transaction *TransactionDTO `json:"transaction"`
}

type ScoreContributionDTO struct {
	Factor        string  `json:"factor"`
	Weight        float64 `json:"weight"`
	Score         float64 `json:"score"`
	WeightedScore float64 `json:"weighted_score"`
}

type AvailabilityDTO struct {
	SupplierID        uuid.UUID `json:"supplier_id"`
	ProductName       string    `json:"product_name"`
	AvailableQuantity float32   `json:"available_quantity"`
}
