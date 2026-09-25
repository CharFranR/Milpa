package dto

import (
	"time"

	"github.com/google/uuid"
)

type ConversationDTO struct {
	ID         uuid.UUID `json:"id"`
	FarmerID   uuid.UUID `json:"farmer_id"`
	BuyerID    uuid.UUID `json:"buyer_id"`
	OfferingID uuid.UUID `json:"offering_id"`
	Visibility bool      `json:"visibility"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type CreateWSConversation struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

type CreateConversationDTO struct {
	ID         uuid.UUID `json:"id"`
	FarmerID   uuid.UUID `json:"farmer_id"`
	BuyerID    uuid.UUID `json:"buyer_id"`
	OfferingID uuid.UUID `json:"offering_id"`
	Visibility bool      `json:"visibility"`
	Now        time.Time `json:"now"`
}

type ResponseConversationDTO struct {
	ID         uuid.UUID `json:"id"`
	FarmerID   uuid.UUID `json:"farmer_id"`
	BuyerID    uuid.UUID `json:"buyer_id"`
	OfferingID uuid.UUID `json:"offering_id"`
	Updated_at time.Time `json:"updated_at"`
}
