package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var testMessageIDs = []uuid.UUID{
	uuid.MustParse("92929292-9292-9292-9292-929292929201"),
	uuid.MustParse("92929292-9292-9292-9292-929292929202"),
	uuid.MustParse("92929292-9292-9292-9292-929292929203"),
	uuid.MustParse("92929292-9292-9292-9292-929292929204"),
	uuid.MustParse("92929292-9292-9292-9292-929292929205"),
	uuid.MustParse("92929292-9292-9292-9292-929292929206"),
}

var testMessageNotFoundID uuid.UUID = uuid.MustParse("92929292-9292-9292-9292-929292929299")

func newMessageFixture(id, conversationID, senderID uuid.UUID, content string, visibility bool, createdAt time.Time) *domain.Message {
	return &domain.Message{
		ID:             id,
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		Visibility:     visibility,
		Created_at:     createdAt,
	}
}

func setupMessageTestData(t *testing.T) {
	t.Helper()
	setupConversationTestData(t)

	conversationRepo := repository.NewConverationImpl(TestPool)

	conversations := []*domain.Conversation{
		newConversationFixture(testConversationID, testConversationBuyerID, testConversationFarmerID, testConversationOfferingID, true, fixedTime),
		newConversationFixture(testConversationID2, testConversationOtherID, testConversationBuyerID, testConversationOfferingID, true, fixedTime.Add(time.Minute)),
	}

	for _, conversation := range conversations {
		if err := conversationRepo.Save(context.Background(), conversation); err != nil {
			t.Fatalf("insert fixture conversation %s: %v", conversation.ID, err)
		}
	}
}

func TestMessageSaveAndGetByID(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewMessageRepositoryImpl(TestPool)

	saved := newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "Is this laptop still available?", true, fixedTime)
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testMessageIDs[0],
			ExpectedErr: nil,
		},
		{
			Name:        "Message Not Found",
			ID:          testMessageNotFoundID,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			message, err := db.GetMessageByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("GetMessageByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("GetMessageByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetMessageByID() unexpected error: %v", err)
				return
			}

			if message == nil {
				t.Fatal("GetMessageByID() returned nil message")
			}
			if message.ID != saved.ID {
				t.Errorf("GetMessageByID() ID = %v, want %v", message.ID, saved.ID)
			}
			if message.ConversationID != saved.ConversationID {
				t.Errorf("GetMessageByID() ConversationID = %v, want %v", message.ConversationID, saved.ConversationID)
			}
			if message.SenderID != saved.SenderID {
				t.Errorf("GetMessageByID() SenderID = %v, want %v", message.SenderID, saved.SenderID)
			}
			if message.Content != saved.Content {
				t.Errorf("GetMessageByID() Content = %v, want %v", message.Content, saved.Content)
			}
			if message.Visibility != saved.Visibility {
				t.Errorf("GetMessageByID() Visibility = %v, want %v", message.Visibility, saved.Visibility)
			}
			if !message.Created_at.Equal(saved.Created_at) {
				t.Errorf("GetMessageByID() Created_at = %v, want %v", message.Created_at, saved.Created_at)
			}
		})
	}
}

func TestMessageBulkSave(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewMessageRepositoryImpl(TestPool)

	batch := []domain.Message{
		*newMessageFixture(testMessageIDs[1], testConversationID, testConversationBuyerID, "Second message", true, fixedTime.Add(time.Minute)),
		*newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "First message", true, fixedTime),
		*newMessageFixture(testMessageIDs[2], testConversationID, testConversationBuyerID, "Third message", true, fixedTime.Add(2*time.Minute)),
	}

	if err := db.BulkSave(context.Background(), &batch); err != nil {
		t.Fatalf("BulkSave() error: %v", err)
	}

	messages, err := db.ListByConversationID(context.Background(), testConversationID)
	if err != nil {
		t.Fatalf("ListByConversationID() error: %v", err)
	}

	expectedIDs := []uuid.UUID{testMessageIDs[0], testMessageIDs[1], testMessageIDs[2]}
	if len(messages) != len(expectedIDs) {
		t.Fatalf("ListByConversationID() got %d messages, want %d", len(messages), len(expectedIDs))
	}
	for i, wantID := range expectedIDs {
		if messages[i].ID != wantID {
			t.Errorf("ListByConversationID()[%d].ID = %v, want %v", i, messages[i].ID, wantID)
		}
	}
}

func TestMessageBulkSaveAtomicity(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewMessageRepositoryImpl(TestPool)

	existing := newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "Already stored", true, fixedTime)
	if err := db.Save(context.Background(), existing); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	batch := []domain.Message{
		*newMessageFixture(testMessageIDs[1], testConversationID, testConversationBuyerID, "Rolled back message", true, fixedTime.Add(time.Minute)),
		*newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "Duplicate primary key", true, fixedTime.Add(2*time.Minute)),
	}

	err := db.BulkSave(context.Background(), &batch)
	if err == nil {
		t.Fatal("BulkSave() with a duplicate primary key: error = nil, want a unique violation")
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("BulkSave() duplicate primary key: error = %v (%T), want *pgconn.PgError", err, err)
	}
	if pgErr.Code != "23505" {
		t.Errorf("BulkSave() duplicate primary key: code = %s, want 23505 (unique_violation)", pgErr.Code)
	}

	_, findErr := db.GetMessageByID(context.Background(), testMessageIDs[1])
	if !errors.Is(findErr, domain.ErrNotFound) {
		t.Errorf("GetMessageByID() for the first message of the failed batch: error = %v, want %v", findErr, domain.ErrNotFound)
	}

	if _, err := db.GetMessageByID(context.Background(), testMessageIDs[0]); err != nil {
		t.Errorf("GetMessageByID() for the pre-existing message: %v", err)
	}
}

func TestMessageListByConversationID(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewMessageRepositoryImpl(TestPool)

	messages := []*domain.Message{
		newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "Later message", true, fixedTime.Add(2*time.Minute)),
		newMessageFixture(testMessageIDs[1], testConversationID, testConversationBuyerID, "Earliest message", true, fixedTime),
		newMessageFixture(testMessageIDs[2], testConversationID, testConversationBuyerID, "Middle message", true, fixedTime.Add(time.Minute)),
		newMessageFixture(testMessageIDs[3], testConversationID, testConversationBuyerID, "Hidden message", false, fixedTime.Add(30*time.Second)),
		newMessageFixture(testMessageIDs[4], testConversationID2, testConversationBuyerID, "Other conversation message", true, fixedTime),
	}

	for _, message := range messages {
		if err := db.Save(context.Background(), message); err != nil {
			t.Fatalf("Save() message %s: %v", message.ID, err)
		}
	}

	tests := []struct {
		Name           string
		ConversationID uuid.UUID
		ExpectedIDs    []uuid.UUID
	}{
		{
			Name:           "Only visible messages of the conversation ordered by created_at ascending",
			ConversationID: testConversationID,
			ExpectedIDs:    []uuid.UUID{testMessageIDs[1], testMessageIDs[2], testMessageIDs[0]},
		},
		{
			Name:           "Same sender in another conversation does not leak in",
			ConversationID: testConversationID2,
			ExpectedIDs:    []uuid.UUID{testMessageIDs[4]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			messages, err := db.ListByConversationID(context.Background(), tt.ConversationID)
			if err != nil {
				t.Errorf("ListByConversationID() unexpected error: %v", err)
				return
			}

			if len(messages) != len(tt.ExpectedIDs) {
				t.Fatalf("ListByConversationID() got %d messages, want %d", len(messages), len(tt.ExpectedIDs))
			}

			for i, wantID := range tt.ExpectedIDs {
				if messages[i].ID != wantID {
					t.Errorf("ListByConversationID()[%d].ID = %v, want %v", i, messages[i].ID, wantID)
				}
			}
		})
	}
}

func TestMessageDelete(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewMessageRepositoryImpl(TestPool)

	saved := newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "To delete", true, fixedTime)
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := db.Delete(context.Background(), testMessageIDs[0]); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetMessageByID(context.Background(), testMessageIDs[0])
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetMessageByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}
