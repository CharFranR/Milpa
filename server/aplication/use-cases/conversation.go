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

type ConversationUseCaseImpl struct {
	conversationRepo port.ConversationRepository
	offeringRepo     port.OfferingRepository
	userRepo         port.UserRepository
	timer            port.TimeProvider
}

func NewConversationUseCase(conversationRepo port.ConversationRepository, offeringRepo port.OfferingRepository, userRepo port.UserRepository, timer port.TimeProvider) *ConversationUseCaseImpl {
	return &ConversationUseCaseImpl{
		conversationRepo: conversationRepo,
		offeringRepo:     offeringRepo,
		userRepo:         userRepo,
		timer:            timer,
	}
}

func (uc *ConversationUseCaseImpl) CreateConversation(ctx context.Context, req dto.CreateConversationDTO) (*dto.ConversationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if req.FarmerID == principal.UserID {
		return nil, domain.ErrInvalidInput
	}

	if _, err := uc.offeringRepo.FindByID(ctx, req.OfferingID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if _, err := uc.userRepo.FindByID(ctx, req.FarmerID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	now := uc.timer.Now()

	conversation, err := domain.NewConvesation(req.FarmerID, principal.UserID, req.OfferingID, now)
	if err != nil {
		return nil, err
	}

	if err := uc.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, err
	}

	return conversationToDTO(conversation), nil
}

func (uc *ConversationUseCaseImpl) ListConversations(ctx context.Context) (*[]dto.ConversationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	conversations, err := uc.conversationRepo.List(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.ConversationDTO, len(conversations))
	for i := range conversations {
		dtos[i] = *conversationToDTO(&conversations[i])
	}

	return &dtos, nil
}

func (uc *ConversationUseCaseImpl) GetConversation(ctx context.Context, id uuid.UUID) (*dto.ConversationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if id == uuid.Nil {
		return nil, fmt.Errorf("Conversation.GetConversation: Null id")
	}

	conversation, err := uc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if !conversation.IsConversationParticipant(principal.UserID) {
		return nil, domain.ErrForbidden
	}

	return conversationToDTO(conversation), nil
}

func (uc *ConversationUseCaseImpl) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	if id == uuid.Nil {
		return fmt.Errorf("Conversation.DeleteConversation: Null id")
	}

	conversation, err := uc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if !conversation.IsConversationParticipant(principal.UserID) {
		return domain.ErrForbidden
	}

	return uc.conversationRepo.Delete(ctx, id)
}

var _ primary.ConversationUserUseCase = (*ConversationUseCaseImpl)(nil)

func conversationToDTO(conversation *domain.Conversation) *dto.ConversationDTO {
	return &dto.ConversationDTO{
		ID:         conversation.ID,
		FarmerID:   conversation.FarmerID,
		BuyerID:    conversation.BuyerID,
		OfferingID: conversation.OfferingID,
		Visibility: conversation.Visibility,
		Created_at: conversation.Created_at,
		Updated_at: conversation.Updated_at,
	}
}
