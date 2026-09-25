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

func TestConversationUseCaseCreateConversation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		req         dto.CreateConversationDTO
		offeringErr error
		userErr     error
		saveErr     error
		wantErr     error
	}{
		{
			name: "happy path",
			ctx:  principalCtx(),
			req:  dto.CreateConversationDTO{FarmerID: testCompanyID, BuyerID: testOtherID, OfferingID: testOfferingID},
		},
		{name: "unauthenticated", ctx: context.Background(), req: dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}, wantErr: auth.ErrUnauthenticated},
		{name: "farmer is principal", ctx: principalCtx(), req: dto.CreateConversationDTO{FarmerID: testUserID, OfferingID: testOfferingID}, wantErr: domain.ErrInvalidInput},
		{name: "offering not found", ctx: principalCtx(), req: dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}, offeringErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "farmer user not found", ctx: principalCtx(), req: dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}, userErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "save error", ctx: principalCtx(), req: dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}, saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conversationRepo := newFakeConversationRepo()
			if tt.saveErr != nil {
				conversationRepo.save = func(ctx context.Context, conversation *domain.Conversation) error {
					return tt.saveErr
				}
			}
			offeringRepo := newFakeOfferingRepo()
			if tt.offeringErr != nil {
				offeringRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
					return nil, tt.offeringErr
				}
			}
			userRepo := newFakeUserRepo()
			if tt.userErr != nil {
				userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, tt.userErr
				}
			}
			uc := usecases.NewConversationUseCase(conversationRepo, offeringRepo, userRepo, newFakeTimer())

			got, err := uc.CreateConversation(tt.ctx, tt.req)

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
			if got.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if got.BuyerID != testUserID {
				t.Errorf("buyer id = %v, want principal %v", got.BuyerID, testUserID)
			}
			if got.FarmerID != tt.req.FarmerID {
				t.Errorf("farmer id = %v, want %v", got.FarmerID, tt.req.FarmerID)
			}
			if got.OfferingID != tt.req.OfferingID {
				t.Errorf("offering id = %v, want %v", got.OfferingID, tt.req.OfferingID)
			}
			if !got.Created_at.Equal(fixedTime) || !got.Updated_at.Equal(fixedTime) {
				t.Errorf("timestamps = %v / %v, want %v", got.Created_at, got.Updated_at, fixedTime)
			}
			if len(conversationRepo.saved) != 1 {
				t.Fatalf("saved conversations = %d, want 1", len(conversationRepo.saved))
			}
			saved := conversationRepo.saved[0]
			if saved.BuyerID != testUserID {
				t.Errorf("saved buyer id = %v, want principal %v", saved.BuyerID, testUserID)
			}
			if saved.FarmerID != tt.req.FarmerID {
				t.Errorf("saved farmer id = %v, want %v", saved.FarmerID, tt.req.FarmerID)
			}
		})
	}
}

func TestConversationUseCaseListConversations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ctx           context.Context
		conversations []domain.Conversation
		listErr       error
		wantLen       int
		wantErr       error
	}{
		{
			name: "happy path",
			ctx:  principalCtx(),
			conversations: func() []domain.Conversation {
				first := *mustConversation()
				second := *mustConversation()
				second.ID = testOtherID
				return []domain.Conversation{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", ctx: principalCtx(), conversations: []domain.Conversation{}, wantLen: 0},
		{name: "unauthenticated", ctx: context.Background(), conversations: []domain.Conversation{}, wantErr: auth.ErrUnauthenticated},
		{name: "repo error", ctx: principalCtx(), listErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conversationRepo := newFakeConversationRepo()
			if tt.listErr != nil {
				conversationRepo.list = func(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
					return nil, tt.listErr
				}
			} else {
				conversationRepo.list = func(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
					return tt.conversations, nil
				}
			}
			uc := usecases.NewConversationUseCase(conversationRepo, newFakeOfferingRepo(), newFakeUserRepo(), newFakeTimer())

			got, err := uc.ListConversations(tt.ctx)

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
			for i, conversation := range *got {
				if conversation.ID != tt.conversations[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, conversation.ID, tt.conversations[i].ID)
				}
				if conversation.BuyerID != tt.conversations[i].BuyerID {
					t.Errorf("dto %d buyer id = %v, want %v", i, conversation.BuyerID, tt.conversations[i].BuyerID)
				}
				if conversation.FarmerID != tt.conversations[i].FarmerID {
					t.Errorf("dto %d farmer id = %v, want %v", i, conversation.FarmerID, tt.conversations[i].FarmerID)
				}
			}
			if len(conversationRepo.listedIDs) != 1 {
				t.Fatalf("list calls = %d, want 1", len(conversationRepo.listedIDs))
			}
			if conversationRepo.listedIDs[0] != testUserID {
				t.Errorf("list user id = %v, want principal %v", conversationRepo.listedIDs[0], testUserID)
			}
		})
	}
}

func TestConversationUseCaseGetConversation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		id         uuid.UUID
		getErr     error
		wantErr    error
		wantAnyErr bool
	}{
		{name: "happy path buyer", ctx: principalCtx(), id: testConversationID},
		{name: "happy path farmer", ctx: principalCtxFor(testCompanyID), id: testConversationID},
		{name: "unauthenticated", ctx: context.Background(), id: testConversationID, wantErr: auth.ErrUnauthenticated},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, wantAnyErr: true},
		{name: "not found", ctx: principalCtx(), id: testConversationID, getErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "not participant", ctx: principalCtxFor(testOtherID), id: testConversationID, wantErr: domain.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conversationRepo := newFakeConversationRepo()
			if tt.getErr != nil {
				conversationRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
					return nil, tt.getErr
				}
			}
			uc := usecases.NewConversationUseCase(conversationRepo, newFakeOfferingRepo(), newFakeUserRepo(), newFakeTimer())

			got, err := uc.GetConversation(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if errors.Is(err, domain.ErrNotFound) {
					t.Fatalf("expected a non-not-found error, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("id = %v, want %v", got.ID, tt.id)
			}
			if got.BuyerID != testUserID {
				t.Errorf("buyer id = %v, want %v", got.BuyerID, testUserID)
			}
			if got.FarmerID != testCompanyID {
				t.Errorf("farmer id = %v, want %v", got.FarmerID, testCompanyID)
			}
			if got.OfferingID != testOfferingID {
				t.Errorf("offering id = %v, want %v", got.OfferingID, testOfferingID)
			}
			if !got.Created_at.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", got.Created_at, fixedTime)
			}
		})
	}
}

func TestConversationUseCaseDeleteConversation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		getErr    error
		deleteErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: testConversationID},
		{name: "unauthenticated", ctx: context.Background(), id: testConversationID, wantErr: auth.ErrUnauthenticated},
		{name: "not found", ctx: principalCtx(), id: testConversationID, getErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "not participant", ctx: principalCtxFor(testOtherID), id: testConversationID, wantErr: domain.ErrForbidden},
		{name: "delete error", ctx: principalCtx(), id: testConversationID, deleteErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conversationRepo := newFakeConversationRepo()
			if tt.getErr != nil {
				conversationRepo.getByID = func(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
					return nil, tt.getErr
				}
			}
			if tt.deleteErr != nil {
				conversationRepo.delete = func(ctx context.Context, id uuid.UUID) error {
					return tt.deleteErr
				}
			}
			uc := usecases.NewConversationUseCase(conversationRepo, newFakeOfferingRepo(), newFakeUserRepo(), newFakeTimer())

			err := uc.DeleteConversation(tt.ctx, tt.id)

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
			if len(conversationRepo.deleted) != 1 {
				t.Fatalf("deleted conversations = %d, want 1", len(conversationRepo.deleted))
			}
			if conversationRepo.deleted[0] != tt.id {
				t.Errorf("deleted id = %v, want %v", conversationRepo.deleted[0], tt.id)
			}
		})
	}
}
