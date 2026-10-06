package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"

	"github.com/google/uuid"
)

type CachedMessageUseCase struct {
	next          primary.MessageUserCase
	cache         port.Cache
	conversations port.ConversationRepository
}

func NewCachedMessageUseCase(next primary.MessageUserCase, cache port.Cache, conversations port.ConversationRepository) *CachedMessageUseCase {
	return &CachedMessageUseCase{
		next:          next,
		cache:         cache,
		conversations: conversations,
	}
}

func (uc *CachedMessageUseCase) CreateMessage(ctx context.Context, req dto.MessageDTO) (*dto.MessageDTO, error) {
	response, err := uc.next.CreateMessage(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "messages:byconversation:"+req.ConversationID.String()+":")
	uc.evictConversationLists(ctx, req.ConversationID)

	return response, nil
}

// evictConversationLists borra el listado cacheado de las dos partes: el preview
// del último mensaje y el contador de no leídos viajan en GET /conversations.
func (uc *CachedMessageUseCase) evictConversationLists(ctx context.Context, conversationID uuid.UUID) {
	conversation, err := uc.conversations.GetByID(ctx, conversationID)
	if err != nil {
		return
	}

	_ = uc.cache.Delete(ctx, "conversations:byuser:"+conversation.BuyerID.String())
	_ = uc.cache.Delete(ctx, "conversations:byuser:"+conversation.FarmerID.String())
}

func (uc *CachedMessageUseCase) ListMessage(ctx context.Context, conversationID uuid.UUID) (*[]dto.MessageDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var messages *[]dto.MessageDTO

	_, err = uc.cache.Remember(
		ctx,
		"messages:byconversation:"+conversationID.String()+":"+principal.UserID.String(),
		5*time.Minute,
		&messages,
		func() error {
			result, err := uc.next.ListMessage(ctx, conversationID)
			if err != nil {
				return err
			}
			messages = result
			return nil
		},
	)

	if err == nil {
		// El use case base acaba de marcar la conversación como leída: la lista
		// cacheada traería el contador anterior.
		_ = uc.cache.Delete(ctx, "conversations:byuser:"+principal.UserID.String())
	}

	return messages, err
}

func (uc *CachedMessageUseCase) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	err := uc.next.DeleteMessage(ctx, messageID)
	if err != nil {
		return err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "messages:")

	return nil
}

var _ primary.MessageUserCase = (*CachedMessageUseCase)(nil)
