package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type MessageUseCaseImpl struct {
	messageRepo      port.MessageRepository
	conversationRepo port.ConversationRepository
	timer            port.TimeProvider
}

func NewMessageUseCase(messageRepo port.MessageRepository, conversationRepo port.ConversationRepository, timer port.TimeProvider) *MessageUseCaseImpl {
	return &MessageUseCaseImpl{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		timer:            timer,
	}
}

func (uc *MessageUseCaseImpl) CreateMessage(ctx context.Context, req dto.MessageDTO) (*dto.MessageDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	now := uc.timer.Now()

	conversation, err := uc.conversationRepo.GetByID(ctx, req.ConversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if !conversation.IsConversationParticipant(principal.UserID) {
		return nil, domain.ErrForbidden
	}

	message, err := domain.NewMessage(req.ConversationID, principal.UserID, req.Content, now)
	if err != nil {
		return nil, err
	}

	return messageToDTO(message), uc.messageRepo.Save(ctx, message)
}

func (uc *MessageUseCaseImpl) ListMessage(ctx context.Context, conversationID uuid.UUID) (*[]dto.MessageDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if conversationID == uuid.Nil {
		return nil, fmt.Errorf("Message.ListMessage: Null conversationID")
	}

	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if !conversation.IsConversationParticipant(principal.UserID) {
		return nil, domain.ErrForbidden
	}

	messages, err := uc.messageRepo.ListByConversationID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.MessageDTO, len(messages))
	for i := range messages {
		dtos[i] = *messageToDTO(&messages[i])
	}

	return &dtos, nil
}

func (uc *MessageUseCaseImpl) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	if messageID == uuid.Nil {
		return fmt.Errorf("Message.DeleteMessage: Null messageID")
	}

	message, err := uc.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if message.SenderID != principal.UserID {
		return domain.ErrForbidden
	}

	return uc.messageRepo.Delete(ctx, messageID)
}

var _ primary.MessageUserCase = (*MessageUseCaseImpl)(nil)

func messageToDTO(message *domain.Message) *dto.MessageDTO {
	return &dto.MessageDTO{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		Visibility:     message.Visibility,
		Created_at:     message.Created_at,
	}
}
