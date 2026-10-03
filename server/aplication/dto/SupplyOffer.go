package dto

import (
	domain "milpa/domain/entities"
	"time"

	"github.com/google/uuid"
)

type SupplyOfferDTO struct {
	ID                  *uuid.UUID                `json:"id"`
	SupplierID          *uuid.UUID                `json:"supplier_id"`
	SupplyRequest       *uuid.UUID                `json:"supply_request_id"`
	TotalAmount         float64                   `json:"total_amount"`
	AmountUnit          domain.MeasurementOptions `json:"measurement"`
	PricePerUnit        *float64                  `json:"price_per_unit"`
	Comments            string                    `json:"comments"`
	ProposedDeliveryDay time.Time                 `json:"delivery_day"`
	DeliveryAvailable   bool                      `json:"delivery_available"`
	Status              domain.OfferStatus        `json:"status"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

type SupplyOfferUpdateDTO struct {
	TotalAmount         float64                   `json:"total_amount"`
	AmountUnit          domain.MeasurementOptions `json:"measurement"`
	PricePerUnit        *float64                  `json:"price_per_unit"`
	Comments            string                    `json:"comments"`
	ProposedDeliveryDay time.Time                 `json:"delivery_day"`
	DeliveryAvailable   bool                      `json:"delivery_available"`
}

type PrioritizedOfferDTO struct {
	Offer             SupplyOfferDTO         `json:"offer"`
	Score             float64                `json:"score"`
	AvailableQuantity float64                `json:"available_quantity"`
	Contributions     []ScoreContributionDTO `json:"contributions"`
}
