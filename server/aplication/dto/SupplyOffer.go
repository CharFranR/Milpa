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
	TotalAmount         float32                   `json:"total_amount"`
	AmountUnit          domain.MeasurementOptions `json:"measurement"`
	ProposedDeliveryDay time.Time                 `json:"delivery_day"`
	DeliveryAvailable   bool                      `json:"delivery_available"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

type SupplyOfferUpdateDTO struct {
	TotalAmount         float32                   `json:"total_amount"`
	AmountUnit          domain.MeasurementOptions `json:"measurement"`
	ProposedDeliveryDay time.Time                 `json:"delivery_day"`
	DeliveryAvailable   bool                      `json:"delivery_available"`
}
