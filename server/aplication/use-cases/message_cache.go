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
	next  primary.MessageUserCase
	cache port.Cache
}

func NewCachedMessageUseCase(next primary.MessageUserCase, cache port.Cache) *CachedMessageUseCase {
	return &CachedMessageUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedMessageUseCase) CreateMessage(ctx context.Context, req dto.MessageDTO) (*dto.MessageDTO, error) {
	response, err := uc.next.CreateMessage(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "messages:byconversation:"+req.ConversationID.String()+":")

	return response, nil
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
