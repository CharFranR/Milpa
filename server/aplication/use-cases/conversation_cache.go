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

type CachedConversationUseCase struct {
	next  primary.ConversationUserUseCase
	cache port.Cache
}

func NewCachedConversationUseCase(next primary.ConversationUserUseCase, cache port.Cache) *CachedConversationUseCase {
	return &CachedConversationUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedConversationUseCase) CreateConversation(ctx context.Context, req dto.CreateConversationDTO) (*dto.ConversationDTO, error) {
	result, err := uc.next.CreateConversation(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.Delete(ctx, "conversations:byuser:"+result.BuyerID.String())

	return result, nil
}

func (uc *CachedConversationUseCase) ListConversations(ctx context.Context) (*[]dto.ConversationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var conversations *[]dto.ConversationDTO

	_, err = uc.cache.Remember(
		ctx,
		"conversations:byuser:"+principal.UserID.String(),
		5*time.Minute,
		&conversations,
		func() error {
			result, err := uc.next.ListConversations(ctx)
			if err != nil {
				return err
			}
			conversations = result
			return nil
		},
	)

	return conversations, err
}

func (uc *CachedConversationUseCase) GetConversation(ctx context.Context, id uuid.UUID) (*dto.ConversationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var conversation *dto.ConversationDTO

	_, err = uc.cache.Remember(
		ctx,
		"conversation:"+id.String()+":"+principal.UserID.String(),
		5*time.Minute,
		&conversation,
		func() error {
			result, err := uc.next.GetConversation(ctx, id)
			if err != nil {
				return err
			}
			conversation = result
			return nil
		},
	)

	return conversation, err
}

func (uc *CachedConversationUseCase) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	err := uc.next.DeleteConversation(ctx, id)
	if err != nil {
		return err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "conversation:"+id.String()+":")

	return nil
}

var _ primary.ConversationUserUseCase = (*CachedConversationUseCase)(nil)
