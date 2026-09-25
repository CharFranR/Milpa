package dto

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
	Visibility     bool      `json:"visibility"`
	Created_at     time.Time `json:"created_at"`
}
