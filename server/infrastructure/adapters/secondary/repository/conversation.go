package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type ConversationRepositoyImpl struct {
	pool DB
}

func NewConverationImpl(pool DB) *ConversationRepositoyImpl {
	return &ConversationRepositoyImpl{
		pool: pool,
	}
}

func (r *ConversationRepositoyImpl) Save(ctx context.Context, conversation *domain.Conversation) error {

	query := `
				INSERT INTO conversations (id, buyer_id, farmer_id, offering_id, visibility, created_at, updated_at)
				VALUES($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query, conversation.ID, conversation.BuyerID, conversation.FarmerID, conversation.OfferingID, conversation.Visibility, conversation.Created_at, conversation.Updated_at)

	if err != nil {
		return fmt.Errorf("Conversation.Save: %w", err)
	}

	return nil
}

func (r *ConversationRepositoyImpl) List(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {

	if userID == uuid.Nil {
		return nil, fmt.Errorf("Conversation.List: Null userID")
	}

	query := `
		SELECT id, buyer_id, farmer_id, offering_id, visibility, created_at, updated_at
		FROM conversations
		WHERE visibility = True AND buyer_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("Conversation.List: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation

	for rows.Next() {

		var conversation domain.Conversation

		if err = rows.Scan(&conversation.ID, &conversation.BuyerID, &conversation.FarmerID, &conversation.OfferingID, &conversation.Visibility, &conversation.Created_at, &conversation.Updated_at); err != nil {
			return nil, fmt.Errorf("Conversation.List: %w", err)
		}

		conversations = append(conversations, conversation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Conversation.List: %w", err)
	}

	return conversations, nil
}

func (r *ConversationRepositoyImpl) ListMessage(ctx context.Context, conversastionID uuid.UUID) ([]domain.Message, error) {

	if conversastionID == uuid.Nil {
		return nil, fmt.Errorf("Conversation.ListMessage: Null conversastionID")
	}

	query := `
		SELECT id, conversation_id, sender_id, content, visibility, created_at
		FROM messages
		WHERE visibility = True AND conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, conversastionID)

	if err != nil {
		return nil, fmt.Errorf("Conversation.ListMessage: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.Content, &message.Visibility, &message.Created_at); err != nil {
			return nil, fmt.Errorf("Conversation.ListMessage: %w", err)
		}

		messages = append(messages, message)

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Conversation.ListMessage: %w", err)
	}

	return messages, nil
}

func (r *ConversationRepositoyImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {

	query := `
		SELECT id, buyer_id, farmer_id, offering_id, visibility, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	var conversation domain.Conversation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&conversation.ID, &conversation.BuyerID, &conversation.FarmerID, &conversation.OfferingID,
		&conversation.Visibility, &conversation.Created_at, &conversation.Updated_at,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Conversation.GetByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("Conversation.GetByID: %w", err)
	}

	return &conversation, nil
}

func (r *ConversationRepositoyImpl) Delete(ctx context.Context, id uuid.UUID) error {

	_, err := r.pool.Exec(ctx, "DELETE FROM conversations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Conversation.Delete: %w", err)
	}

	return nil
}

var _ port.ConversationRepository = (*ConversationRepositoyImpl)(nil)
