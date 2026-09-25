package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testConversationBuyerID uuid.UUID = uuid.MustParse("15151515-1515-1515-1515-151515151515")
var testConversationFarmerID uuid.UUID = uuid.MustParse("16161616-1616-1616-1616-161616161616")
var testConversationOtherID uuid.UUID = uuid.MustParse("17171717-1717-1717-1717-171717171717")
var testConversationOfferingID uuid.UUID = uuid.MustParse("45454545-4545-4545-4545-454545454545")

var testConversationID uuid.UUID = uuid.MustParse("91919191-9191-9191-9191-919191919101")
var testConversationID2 uuid.UUID = uuid.MustParse("91919191-9191-9191-9191-919191919102")
var testConversationID3 uuid.UUID = uuid.MustParse("91919191-9191-9191-9191-919191919103")
var testConversationID4 uuid.UUID = uuid.MustParse("91919191-9191-9191-9191-919191919104")
var testConversationNotFoundID uuid.UUID = uuid.MustParse("91919191-9191-9191-9191-919191919199")

func newConversationFixture(id, buyerID, farmerID, offeringID uuid.UUID, visibility bool, createdAt time.Time) *domain.Conversation {
	return &domain.Conversation{
		ID:         id,
		BuyerID:    buyerID,
		FarmerID:   farmerID,
		OfferingID: offeringID,
		Visibility: visibility,
		Created_at: createdAt,
		Updated_at: createdAt,
	}
}

func setupConversationTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)
	offeringRepo := repository.NewOfferingRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testConversationBuyerID,
			FirstName:    "Conversation",
			LastName:     "Buyer",
			Role:         domain.RoleMIPYME,
			Email:        "conversation-buyer@example.com",
			PhoneNumber:  "1111-1111",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testConversationFarmerID,
			FirstName:    "Conversation",
			LastName:     "Farmer",
			Role:         domain.RoleMIPYME,
			Email:        "conversation-farmer@example.com",
			PhoneNumber:  "2222-2222",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testConversationOtherID,
			FirstName:    "Conversation",
			LastName:     "Other",
			Role:         domain.RoleMIPYME,
			Email:        "conversation-other@example.com",
			PhoneNumber:  "3333-3333",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}

	for _, u := range users {
		if _, err := userRepo.Save(context.Background(), u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}

	err := offeringRepo.Save(context.Background(), &domain.Offering{
		ID:        testConversationOfferingID,
		UserID:    testConversationFarmerID,
		Type:      domain.OfferingProduct,
		Name:      "Conversation Offering",
		Price:     100.00,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	})
	if err != nil {
		t.Fatalf("insert offering: %v", err)
	}
}

func TestConversationSaveAndGetByID(t *testing.T) {
	setupConversationTestData(t)
	db := repository.NewConverationImpl(TestPool)

	saved := newConversationFixture(testConversationID, testConversationBuyerID, testConversationFarmerID, testConversationOfferingID, true, fixedTime)
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
			ID:          testConversationID,
			ExpectedErr: nil,
		},
		{
			Name:        "Conversation Not Found",
			ID:          testConversationNotFoundID,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			conversation, err := db.GetByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("GetByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("GetByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error: %v", err)
				return
			}

			if conversation == nil {
				t.Fatal("GetByID() returned nil conversation")
			}
			if conversation.ID != saved.ID {
				t.Errorf("GetByID() ID = %v, want %v", conversation.ID, saved.ID)
			}
			if conversation.BuyerID != saved.BuyerID {
				t.Errorf("GetByID() BuyerID = %v, want %v", conversation.BuyerID, saved.BuyerID)
			}
			if conversation.FarmerID != saved.FarmerID {
				t.Errorf("GetByID() FarmerID = %v, want %v", conversation.FarmerID, saved.FarmerID)
			}
			if conversation.OfferingID != saved.OfferingID {
				t.Errorf("GetByID() OfferingID = %v, want %v", conversation.OfferingID, saved.OfferingID)
			}
			if conversation.Visibility != saved.Visibility {
				t.Errorf("GetByID() Visibility = %v, want %v", conversation.Visibility, saved.Visibility)
			}
			if !conversation.Created_at.Equal(saved.Created_at) {
				t.Errorf("GetByID() Created_at = %v, want %v", conversation.Created_at, saved.Created_at)
			}
			if !conversation.Updated_at.Equal(saved.Updated_at) {
				t.Errorf("GetByID() Updated_at = %v, want %v", conversation.Updated_at, saved.Updated_at)
			}
		})
	}
}

func TestConversationList(t *testing.T) {
	setupConversationTestData(t)
	db := repository.NewConverationImpl(TestPool)

	fixtures := []*domain.Conversation{
		newConversationFixture(testConversationID, testConversationBuyerID, testConversationFarmerID, testConversationOfferingID, true, fixedTime),
		newConversationFixture(testConversationID2, testConversationOtherID, testConversationBuyerID, testConversationOfferingID, true, fixedTime.Add(time.Minute)),
		newConversationFixture(testConversationID3, testConversationBuyerID, testConversationFarmerID, testConversationOfferingID, false, fixedTime.Add(2*time.Minute)),
		newConversationFixture(testConversationID4, testConversationOtherID, testConversationFarmerID, testConversationOfferingID, true, fixedTime.Add(3*time.Minute)),
	}

	for _, conversation := range fixtures {
		if err := db.Save(context.Background(), conversation); err != nil {
			t.Fatalf("Save() conversation %s: %v", conversation.ID, err)
		}
	}

	tests := []struct {
		Name          string
		UserID        uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Buyer sees only visible conversations where they are the buyer",
			UserID:        testConversationBuyerID,
			ExpectedIDs:   []uuid.UUID{testConversationID},
			ExpectedTotal: 1,
		},
		{
			Name:          "Farmer of the conversations does not appear in the buyer list",
			UserID:        testConversationFarmerID,
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
		{
			Name:          "Other buyer sees both of their conversations including the shared farmer one",
			UserID:        testConversationOtherID,
			ExpectedIDs:   []uuid.UUID{testConversationID2, testConversationID4},
			ExpectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			conversations, err := db.List(context.Background(), tt.UserID)
			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
				return
			}

			if len(conversations) != tt.ExpectedTotal {
				t.Fatalf("List() got %d conversations, want %d", len(conversations), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(conversations))
			for _, conversation := range conversations {
				found[conversation.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("List() did not return conversation %v", wantID)
				}
			}
		})
	}
}

func TestConversationListMessage(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewConverationImpl(TestPool)
	messageRepo := repository.NewMessageRepositoryImpl(TestPool)

	messages := []*domain.Message{
		newMessageFixture(testMessageIDs[0], testConversationID, testConversationBuyerID, "Later message", true, fixedTime.Add(2*time.Minute)),
		newMessageFixture(testMessageIDs[1], testConversationID, testConversationBuyerID, "Earliest message", true, fixedTime),
		newMessageFixture(testMessageIDs[2], testConversationID, testConversationBuyerID, "Middle message", true, fixedTime.Add(time.Minute)),
		newMessageFixture(testMessageIDs[3], testConversationID, testConversationBuyerID, "Hidden message", false, fixedTime.Add(30*time.Second)),
		newMessageFixture(testMessageIDs[4], testConversationID2, testConversationBuyerID, "Other conversation message", true, fixedTime),
	}

	for _, message := range messages {
		if err := messageRepo.Save(context.Background(), message); err != nil {
			t.Fatalf("Save() message %s: %v", message.ID, err)
		}
	}

	tests := []struct {
		Name           string
		ConversationID uuid.UUID
		ExpectedIDs    []uuid.UUID
	}{
		{
			Name:           "Only the messages of the conversation ordered by created_at ascending",
			ConversationID: testConversationID,
			ExpectedIDs:    []uuid.UUID{testMessageIDs[1], testMessageIDs[2], testMessageIDs[0]},
		},
		{
			Name:           "Messages of the other conversation with the same sender",
			ConversationID: testConversationID2,
			ExpectedIDs:    []uuid.UUID{testMessageIDs[4]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			messages, err := db.ListMessage(context.Background(), tt.ConversationID)
			if err != nil {
				t.Errorf("ListMessage() unexpected error: %v", err)
				return
			}

			if len(messages) != len(tt.ExpectedIDs) {
				t.Fatalf("ListMessage() got %d messages, want %d", len(messages), len(tt.ExpectedIDs))
			}

			for i, wantID := range tt.ExpectedIDs {
				if messages[i].ID != wantID {
					t.Errorf("ListMessage()[%d].ID = %v, want %v", i, messages[i].ID, wantID)
				}
			}
		})
	}
}

func TestConversationDelete(t *testing.T) {
	setupMessageTestData(t)
	db := repository.NewConverationImpl(TestPool)
	messageRepo := repository.NewMessageRepositoryImpl(TestPool)

	for i, content := range []string{"First message", "Second message"} {
		message := newMessageFixture(testMessageIDs[i], testConversationID, testConversationBuyerID, content, true, fixedTime.Add(time.Duration(i)*time.Minute))
		if err := messageRepo.Save(context.Background(), message); err != nil {
			t.Fatalf("Save() message %s: %v", message.ID, err)
		}
	}

	if err := db.Delete(context.Background(), testConversationID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetByID(context.Background(), testConversationID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}

	remaining, err := messageRepo.ListByConversationID(context.Background(), testConversationID)
	if err != nil {
		t.Fatalf("ListByConversationID() after Delete() error: %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("ListByConversationID() after Delete() got %d messages, want 0 (cascade)", len(remaining))
	}
}

func TestConversationMessageIndexesExist(t *testing.T) {
	ctx := context.Background()

	indexes := []string{
		"idx_conversations_buyer",
		"idx_conversations_farmer",
		"idx_conversations_offering",
		"idx_messages_conversation_created",
	}

	for _, name := range indexes {
		t.Run(name, func(t *testing.T) {
			var found bool
			err := TestPool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)", name).Scan(&found)
			if err != nil {
				t.Fatalf("query pg_indexes: %v", err)
			}
			if !found {
				t.Errorf("index %s does not exist", name)
			}
		})
	}
}
