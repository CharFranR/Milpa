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

type MessageRepositoryImpl struct {
	pool DB
}

func NewMessageRepositoryImpl(pool DB) *MessageRepositoryImpl {
	return &MessageRepositoryImpl{
		pool: pool,
	}
}

func (r *MessageRepositoryImpl) Save(ctx context.Context, message *domain.Message) error {

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content, visibility, created_at)
		VALUES($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query, message.ID, message.ConversationID, message.SenderID, message.Content, message.Visibility, message.Created_at)

	if err != nil {
		return fmt.Errorf("Message.Save: %w", err)
	}

	return nil
}

func (r *MessageRepositoryImpl) BulkSave(ctx context.Context, messages *[]domain.Message) error {

	if messages == nil || len(*messages) == 0 {
		return nil
	}

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content, visibility, created_at)
		VALUES($1, $2, $3, $4, $5, $6)
	`

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Message.BulkSave: %w", err)
	}

	defer tx.Rollback(ctx)

	for _, message := range *messages {
		_, err := tx.Exec(ctx, query, message.ID, message.ConversationID, message.SenderID, message.Content, message.Visibility, message.Created_at)
		if err != nil {
			return fmt.Errorf("Message.BulkSave: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Message.BulkSave: %w", err)
	}

	return nil
}

func (r *MessageRepositoryImpl) ListByConversationID(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error) {

	if conversationID == uuid.Nil {
		return nil, fmt.Errorf("Message.ListByConversationID: Null conversationID")
	}

	query := `
		SELECT id, conversation_id, sender_id, content, visibility, created_at
		FROM messages
		WHERE visibility = True AND conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, conversationID)

	if err != nil {
		return nil, fmt.Errorf("Message.ListByConversationID: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.Content, &message.Visibility, &message.Created_at); err != nil {
			return nil, fmt.Errorf("Message.ListByConversationID: %w", err)
		}

		messages = append(messages, message)

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Message.ListByConversationID: %w", err)
	}

	return messages, nil
}

func (r *MessageRepositoryImpl) GetMessageByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {

	query := `
		SELECT id, conversation_id, sender_id, content, visibility, created_at
		FROM messages
		WHERE id = $1
	`

	var message domain.Message
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&message.ID, &message.ConversationID, &message.SenderID, &message.Content,
		&message.Visibility, &message.Created_at,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Message.GetMessageByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("Message.GetMessageByID: %w", err)
	}

	return &message, nil
}

func (r *MessageRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {

	_, err := r.pool.Exec(ctx, "DELETE FROM messages WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Message.Delete: %w", err)
	}

	return nil
}

var _ port.MessageRepository = (*MessageRepositoryImpl)(nil)
