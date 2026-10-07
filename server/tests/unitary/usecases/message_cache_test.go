package usecases_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
)

// The inbox list carries the last-message preview and the unread counter, so a
// new message has to evict both participants: otherwise the recipient watches a
// frozen preview for up to five minutes and never sees the badge.
func TestMessageCacheEvictsInboxesOnCreate(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	conversations := newFakeConversationRepo()
	uc := usecases.NewCachedMessageUseCase(
		usecases.NewMessageUseCase(newFakeMessageRepo(), conversations, newFakeTimer()),
		cache,
		conversations,
	)

	if _, err := uc.CreateMessage(principalCtx(), dto.MessageDTO{ConversationID: testConversationID, Content: "¿Todavía hay?"}); err != nil {
		t.Fatalf("CreateMessage() error: %v", err)
	}

	wantPrefix := "messages:byconversation:" + testConversationID.String() + ":"
	if cache.deletedPrefix != wantPrefix {
		t.Errorf("deleted prefix = %q, want %q", cache.deletedPrefix, wantPrefix)
	}

	conversation := mustConversation()
	for _, want := range []string{
		"conversations:byuser:" + conversation.BuyerID.String(),
		"conversations:byuser:" + conversation.FarmerID.String(),
	} {
		if !slices.Contains(cache.deletedKeys, want) {
			t.Errorf("deleted keys = %v, want %q", cache.deletedKeys, want)
		}
	}
}

// Opening the chat marks it read, so the caller's own inbox entry has to be
// recomputed: the cached one would still say "3 unread".
func TestMessageCacheEvictsOwnInboxOnList(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	conversations := newFakeConversationRepo()
	uc := usecases.NewCachedMessageUseCase(
		usecases.NewMessageUseCase(newFakeMessageRepo(), conversations, newFakeTimer()),
		cache,
		conversations,
	)

	if _, err := uc.ListMessage(principalCtx(), testConversationID); err != nil {
		t.Fatalf("ListMessage() error: %v", err)
	}

	want := "conversations:byuser:" + testUserID.String()
	if !slices.Contains(cache.deletedKeys, want) {
		t.Errorf("deleted keys = %v, want %q", cache.deletedKeys, want)
	}
}

func TestMessageUseCaseListMarksConversationRead(t *testing.T) {
	t.Parallel()

	conversations := newFakeConversationRepo()
	uc := usecases.NewMessageUseCase(newFakeMessageRepo(), conversations, newFakeTimer())

	if _, err := uc.ListMessage(principalCtx(), testConversationID); err != nil {
		t.Fatalf("ListMessage() error: %v", err)
	}

	if !slices.Contains(conversations.markedRead, testUserID) {
		t.Errorf("marked read for = %v, want %v", conversations.markedRead, testUserID)
	}
}

// A read receipt is a nicety: failing to stamp it must not cost the user the
// history they asked for.
func TestMessageUseCaseListSurvivesMarkReadFailure(t *testing.T) {
	t.Parallel()

	conversations := newFakeConversationRepo()
	conversations.markRead = func(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID, at time.Time) (bool, error) {
		return false, errFake
	}
	uc := usecases.NewMessageUseCase(newFakeMessageRepo(), conversations, newFakeTimer())

	messages, err := uc.ListMessage(principalCtx(), testConversationID)
	if err != nil {
		t.Fatalf("ListMessage() error: %v", err)
	}
	if len(*messages) != 1 {
		t.Errorf("messages = %d, want 1", len(*messages))
	}
}
