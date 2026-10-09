package primary

import (
	"context"
	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type MessageUserCase interface {
	CreateMessage(ctx context.Context, req dto.MessageDTO) (*dto.MessageDTO, error)
	// BulkCreateMessage(ctx context.Context, req []dto.MessageDTO) error, por el momento no vamos a menjar esta implementación
	ListMessage(ctx context.Context, conversationID uuid.UUID) (*[]dto.MessageDTO, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
}
