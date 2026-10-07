package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
			INSERT INTO conversations (id, buyer_id, farmer_id, offering_id, match_id, visibility, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query, conversation.ID, conversation.BuyerID, conversation.FarmerID, nullUUID(conversation.OfferingID), conversation.MatchID, conversation.Visibility, conversation.Created_at, conversation.Updated_at)

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
		SELECT c.id, c.buyer_id, c.farmer_id, c.offering_id, c.match_id, c.visibility, c.created_at, c.updated_at,
			last.content, last.sender_id, last.created_at,
			(
				SELECT COUNT(*) FROM messages m
				WHERE m.conversation_id = c.id
					AND m.visibility = True
					AND m.sender_id <> $1
					AND m.created_at > COALESCE(
						CASE WHEN c.buyer_id = $1 THEN c.buyer_last_read_at END,
						CASE WHEN c.farmer_id = $1 THEN c.farmer_last_read_at END,
						to_timestamp(0)
					)
			)
		FROM conversations c
		LEFT JOIN LATERAL (
			SELECT m.content, m.sender_id, m.created_at
			FROM messages m
			WHERE m.conversation_id = c.id AND m.visibility = True
			ORDER BY m.created_at DESC
			LIMIT 1
		) last ON True
		WHERE c.visibility = True AND (c.buyer_id = $1 OR c.farmer_id = $1)
	`

	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("Conversation.List: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation

	for rows.Next() {

		var conversation domain.Conversation
		var offeringID, matchID *uuid.UUID
		var lastContent *string
		var lastSender *uuid.UUID
		var lastCreatedAt *time.Time

		if err = rows.Scan(
			&conversation.ID, &conversation.BuyerID, &conversation.FarmerID, &offeringID, &matchID,
			&conversation.Visibility, &conversation.Created_at, &conversation.Updated_at,
			&lastContent, &lastSender, &lastCreatedAt, &conversation.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("Conversation.List: %w", err)
		}
		if offeringID != nil {
			conversation.OfferingID = *offeringID
		}
		conversation.MatchID = matchID

		if lastContent != nil && lastSender != nil && lastCreatedAt != nil {
			conversation.LastMessage = &domain.LastMessage{
				Content:    *lastContent,
				SenderID:   *lastSender,
				Created_at: *lastCreatedAt,
			}
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
		SELECT id, buyer_id, farmer_id, offering_id, match_id, visibility, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	var conversation domain.Conversation
	var offeringID, matchID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&conversation.ID, &conversation.BuyerID, &conversation.FarmerID, &offeringID, &matchID,
		&conversation.Visibility, &conversation.Created_at, &conversation.Updated_at,
	)
	if offeringID != nil {
		conversation.OfferingID = *offeringID
	}
	conversation.MatchID = matchID

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Conversation.GetByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("Conversation.GetByID: %w", err)
	}

	return &conversation, nil
}

func (r *ConversationRepositoyImpl) MarkRead(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID, at time.Time) (bool, error) {

	if conversationID == uuid.Nil || userID == uuid.Nil {
		return false, fmt.Errorf("Conversation.MarkRead: Null conversationID or userID")
	}

	// Cada parte tiene su propia marca: abrir el chat como comprador no puede
	// borrar los no leídos del agricultor.
	query := `
		UPDATE conversations
		SET buyer_last_read_at = CASE WHEN buyer_id = $2 THEN $3 ELSE buyer_last_read_at END,
			farmer_last_read_at = CASE WHEN farmer_id = $2 THEN $3 ELSE farmer_last_read_at END
		WHERE id = $1 AND (buyer_id = $2 OR farmer_id = $2)
	`

	tag, err := r.pool.Exec(ctx, query, conversationID, userID, at)
	if err != nil {
		return false, fmt.Errorf("Conversation.MarkRead: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (r *ConversationRepositoyImpl) Delete(ctx context.Context, id uuid.UUID) error {

	_, err := r.pool.Exec(ctx, "DELETE FROM conversations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Conversation.Delete: %w", err)
	}

	return nil
}

var _ port.ConversationRepository = (*ConversationRepositoyImpl)(nil)
