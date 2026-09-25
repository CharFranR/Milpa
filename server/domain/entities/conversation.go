package domain

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID         uuid.UUID
	FarmerID   uuid.UUID
	BuyerID    uuid.UUID
	OfferingID uuid.UUID
	Visibility bool
	Created_at time.Time
	Updated_at time.Time
}

func NewConvesation(FamerID uuid.UUID, BuyerID uuid.UUID, OfferingID uuid.UUID, now time.Time) (*Conversation, error) {
	if FamerID == uuid.Nil {
		return nil, ErrFarmerConversationRequired
	}

	if BuyerID == uuid.Nil {
		return nil, ErrBuyerConversationRequired
	}

	if OfferingID == uuid.Nil {
		return nil, ErrOfferingconversationRequied
	}

	return &Conversation{
		ID:         uuid.New(),
		FarmerID:   FamerID,
		BuyerID:    BuyerID,
		OfferingID: OfferingID,
		Visibility: true,
		Created_at: now,
		Updated_at: now,
	}, nil
}
