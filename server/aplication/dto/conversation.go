package dto

import (
	"time"

	"github.com/google/uuid"
)

type ConversationDTO struct {
	ID         uuid.UUID  `json:"id"`
	FarmerID   uuid.UUID  `json:"farmer_id"`
	BuyerID    uuid.UUID  `json:"buyer_id"`
	OfferingID uuid.UUID  `json:"offering_id"`
	MatchID    *uuid.UUID `json:"match_id,omitempty"`
	Visibility bool       `json:"visibility"`
	Created_at time.Time  `json:"created_at"`
	Updated_at time.Time  `json:"updated_at"`
	// Los llena solo el listado: el resto de endpoints no los consulta y los
	// deja en cero/nil.
	LastMessage *LastMessageDTO `json:"last_message,omitempty"`
	UnreadCount int             `json:"unread_count"`
}

type LastMessageDTO struct {
	Content    string    `json:"content"`
	SenderID   uuid.UUID `json:"sender_id"`
	Created_at time.Time `json:"created_at"`
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
