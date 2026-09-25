package primary

import (
	"context"
	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type ConversationUserUseCase interface {
	CreateConversation(ctx context.Context, req dto.CreateConversationDTO) (*dto.ConversationDTO, error)
	ListConversations(ctx context.Context) (*[]dto.ConversationDTO, error)
	GetConversation(ctx context.Context, req uuid.UUID) (*dto.ConversationDTO, error)
	DeleteConversation(ctx context.Context, req uuid.UUID) error
}
