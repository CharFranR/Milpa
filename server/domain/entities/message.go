package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
	Visibility     bool
	Created_at     time.Time
}

func NewMessage(ConversationID uuid.UUID, SenderID uuid.UUID, Content string, now time.Time) (*Message, error) {

	if ConversationID == uuid.Nil {
		return nil, ErrConversationRequired
	}

	if SenderID == uuid.Nil {
		return nil, ErrSenderMessageRequired
	}

	if Content == "" {
		return nil, ErrContentMessageRequired
	}

	return &Message{
		ID:             uuid.New(),
		ConversationID: ConversationID,
		SenderID:       SenderID,
		Content:        Content,
		Visibility:     true,
		Created_at:     now,
	}, nil

}
