package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestMessageUseCaseCreateMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		req     dto.MessageDTO
		getErr  error
		saveErr error
		wantErr error
	}{
		{
			name: "happy path",
			ctx:  principalCtx(),
			req:  dto.MessageDTO{ConversationID: testConversationID, SenderID: testOtherID, Content: "Is it still available?"},
		},
		{name: "unauthenticated", ctx: context.Background(), req: dto.MessageDTO{ConversationID: testConversationID, Content: "Is it still available?"}, wantErr: auth.ErrUnauthenticated},
		{name: "conversation not found", ctx: principalCtx(), req: dto.MessageDTO{ConversationID: testConversationID, Content: "Is it still available?"}, getErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "not participant", ctx: principalCtxFor(testOtherID), req: dto.MessageDTO{ConversationID: testConversationID, Content: "Is it still available?"}, wantErr: domain.ErrForbidden},
		{name: "empty content", ctx: principalCtx(), req: dto.MessageDTO{ConversationID: testConversationID}, wantErr: domain.ErrContentMessageRequired},
		{name: "save error", ctx: principalCtx(), req: dto.MessageDTO{ConversationID: testConversationID, Content: "Is it still available?"}, saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messageRepo := newFakeMessageRepo()
			if tt.saveErr != nil {
				messageRepo.save = func(ctx context.Context, message *domain.Message) error {
					return tt.saveErr
				}
			}
			conversationRepo := newFakeConversationRepo()
			if tt.getErr != nil {
				conversationRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
					return nil, tt.getErr
				}
			}
			uc := usecases.NewMessageUseCase(messageRepo, conversationRepo, newFakeTimer())

			err := uc.CreateMessage(tt.ctx, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(messageRepo.saved) != 1 {
				t.Fatalf("saved messages = %d, want 1", len(messageRepo.saved))
			}
			saved := messageRepo.saved[0]
			if saved.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if saved.SenderID != testUserID {
				t.Errorf("sender id = %v, want principal %v", saved.SenderID, testUserID)
			}
			if saved.ConversationID != tt.req.ConversationID {
				t.Errorf("conversation id = %v, want %v", saved.ConversationID, tt.req.ConversationID)
			}
			if saved.Content != tt.req.Content {
				t.Errorf("content = %q, want %q", saved.Content, tt.req.Content)
			}
			if !saved.Created_at.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", saved.Created_at, fixedTime)
			}
		})
	}
}

func TestMessageUseCaseListMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ctx      context.Context
		id       uuid.UUID
		getErr   error
		listErr  error
		messages []domain.Message
		wantLen  int
		wantErr  error
	}{
		{
			name: "happy path",
			ctx:  principalCtx(),
			id:   testConversationID,
			messages: func() []domain.Message {
				first := *mustMessage()
				second := *mustMessage()
				second.ID = testOtherID
				second.Content = "Yes, it is available"
				return []domain.Message{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", ctx: principalCtx(), id: testConversationID, messages: []domain.Message{}, wantLen: 0},
		{name: "unauthenticated", ctx: context.Background(), id: testConversationID, wantErr: auth.ErrUnauthenticated},
		{name: "conversation not found", ctx: principalCtx(), id: testConversationID, getErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "not participant", ctx: principalCtxFor(testOtherID), id: testConversationID, wantErr: domain.ErrForbidden},
		{name: "repo error", ctx: principalCtx(), id: testConversationID, listErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messageRepo := newFakeMessageRepo()
			if tt.listErr != nil {
				messageRepo.listByConversationID = func(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error) {
					return nil, tt.listErr
				}
			} else {
				messageRepo.listByConversationID = func(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error) {
					return tt.messages, nil
				}
			}
			conversationRepo := newFakeConversationRepo()
			if tt.getErr != nil {
				conversationRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
					return nil, tt.getErr
				}
			}
			uc := usecases.NewMessageUseCase(messageRepo, conversationRepo, newFakeTimer())

			got, err := uc.ListMessage(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected a non-nil slice pointer, got nil")
			}
			if len(*got) != tt.wantLen {
				t.Fatalf("dtos = %d, want %d", len(*got), tt.wantLen)
			}
			for i, message := range *got {
				if message.ID != tt.messages[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, message.ID, tt.messages[i].ID)
				}
				if message.ConversationID != tt.messages[i].ConversationID {
					t.Errorf("dto %d conversation id = %v, want %v", i, message.ConversationID, tt.messages[i].ConversationID)
				}
				if message.SenderID != tt.messages[i].SenderID {
					t.Errorf("dto %d sender id = %v, want %v", i, message.SenderID, tt.messages[i].SenderID)
				}
				if message.Content != tt.messages[i].Content {
					t.Errorf("dto %d content = %q, want %q", i, message.Content, tt.messages[i].Content)
				}
			}
		})
	}
}

func TestMessageUseCaseDeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		getErr    error
		deleteErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: testMessageID},
		{name: "unauthenticated", ctx: context.Background(), id: testMessageID, wantErr: auth.ErrUnauthenticated},
		{name: "message not found", ctx: principalCtx(), id: testMessageID, getErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "sender mismatch", ctx: principalCtxFor(testOtherID), id: testMessageID, wantErr: domain.ErrForbidden},
		{name: "delete error", ctx: principalCtx(), id: testMessageID, deleteErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messageRepo := newFakeMessageRepo()
			if tt.getErr != nil {
				messageRepo.getMessageByID = func(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
					return nil, tt.getErr
				}
			}
			if tt.deleteErr != nil {
				messageRepo.delete = func(ctx context.Context, id uuid.UUID) error {
					return tt.deleteErr
				}
			}
			uc := usecases.NewMessageUseCase(messageRepo, newFakeConversationRepo(), newFakeTimer())

			err := uc.DeleteMessage(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(messageRepo.deleted) != 1 {
				t.Fatalf("deleted messages = %d, want 1", len(messageRepo.deleted))
			}
			if messageRepo.deleted[0] != tt.id {
				t.Errorf("deleted id = %v, want %v", messageRepo.deleted[0], tt.id)
			}
		})
	}
}
