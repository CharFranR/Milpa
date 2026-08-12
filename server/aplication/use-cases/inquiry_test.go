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

func TestInquiryUseCaseCreateInquiry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		offeringID uuid.UUID
		message    string
		saveErr    error
		wantErr    error
	}{
		{name: "happy path", ctx: principalCtx(), offeringID: testOfferingID, message: "Is it in stock?"},
		{name: "unauthenticated", ctx: context.Background(), offeringID: testOfferingID, message: "Is it in stock?", wantErr: auth.ErrUnauthenticated},
		{name: "empty message", ctx: principalCtx(), offeringID: testOfferingID, wantErr: domain.ErrMessageRequired},
		{name: "save error", ctx: principalCtx(), offeringID: testOfferingID, message: "Is it in stock?", saveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inquiryRepo := newFakeInquiryRepo()
			if tt.saveErr != nil {
				inquiryRepo.save = func(ctx context.Context, inquiry *domain.Inquiry) error {
					return tt.saveErr
				}
			}
			uc := usecases.NewInquiryUseCase(inquiryRepo, newFakeTimer())

			got, err := uc.CreateInquiry(tt.ctx, dto.CreateInquiryRequest{OfferingID: tt.offeringID, Message: tt.message})

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
			if got.UserID != testUserID {
				t.Errorf("user id = %v, want %v", got.UserID, testUserID)
			}
			if got.OfferingID != tt.offeringID {
				t.Errorf("offering id = %v, want %v", got.OfferingID, tt.offeringID)
			}
			if got.Message != tt.message {
				t.Errorf("message = %q, want %q", got.Message, tt.message)
			}
			if got.Status != domain.InquiryPending {
				t.Errorf("status = %v, want %v", got.Status, domain.InquiryPending)
			}
			if !got.CreatedAt.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", got.CreatedAt, fixedTime)
			}
			if !got.UpdatedAt.Equal(got.CreatedAt) {
				t.Errorf("updated at = %v, want %v", got.UpdatedAt, got.CreatedAt)
			}
			if len(inquiryRepo.saved) != 1 {
				t.Fatalf("saved inquiries = %d, want 1", len(inquiryRepo.saved))
			}
			saved := inquiryRepo.saved[0]
			if saved.UserID != testUserID {
				t.Errorf("saved user id = %v, want %v", saved.UserID, testUserID)
			}
			if saved.Message != tt.message {
				t.Errorf("saved message = %q, want %q", saved.Message, tt.message)
			}
			if saved.Status != domain.InquiryPending {
				t.Errorf("saved status = %v, want %v", saved.Status, domain.InquiryPending)
			}
		})
	}
}

func TestInquiryUseCaseGetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "happy path"},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
		{name: "not found", repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inquiryRepo := newFakeInquiryRepo()
			if tt.repoErr != nil {
				inquiryRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewInquiryUseCase(inquiryRepo, newFakeTimer())

			got, err := uc.GetByID(context.Background(), testInquiryID)

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
			if got.ID != testInquiryID {
				t.Errorf("id = %v, want %v", got.ID, testInquiryID)
			}
			if got.UserID != testUserID {
				t.Errorf("user id = %v, want %v", got.UserID, testUserID)
			}
			if got.OfferingID != testOfferingID {
				t.Errorf("offering id = %v, want %v", got.OfferingID, testOfferingID)
			}
			if got.Message != "Is it in stock?" {
				t.Errorf("message = %q, want %q", got.Message, "Is it in stock?")
			}
			if got.Status != domain.InquiryPending {
				t.Errorf("status = %v, want %v", got.Status, domain.InquiryPending)
			}
			if !got.CreatedAt.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", got.CreatedAt, fixedTime)
			}
			if !got.UpdatedAt.Equal(got.CreatedAt) {
				t.Errorf("updated at = %v, want %v", got.UpdatedAt, got.CreatedAt)
			}
		})
	}
}

func TestInquiryUseCaseGetByUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inquiries []domain.Inquiry
		repoErr   error
		wantLen   int
		wantErr   error
	}{
		{
			name: "happy path",
			inquiries: func() []domain.Inquiry {
				first := *mustInquiry()
				second := *mustInquiry()
				second.ID = testOtherID
				second.Message = "Do you deliver?"
				return []domain.Inquiry{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", inquiries: []domain.Inquiry{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inquiryRepo := newFakeInquiryRepo()
			if tt.repoErr != nil {
				inquiryRepo.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error) {
					return nil, tt.repoErr
				}
			} else {
				inquiryRepo.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error) {
					return tt.inquiries, nil
				}
			}
			uc := usecases.NewInquiryUseCase(inquiryRepo, newFakeTimer())

			got, err := uc.GetByUser(context.Background(), testUserID)

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
			if len(got) != tt.wantLen {
				t.Fatalf("dtos = %d, want %d", len(got), tt.wantLen)
			}
			for i, dto := range got {
				if dto.ID != tt.inquiries[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.inquiries[i].ID)
				}
				if dto.Message != tt.inquiries[i].Message {
					t.Errorf("dto %d message = %q, want %q", i, dto.Message, tt.inquiries[i].Message)
				}
				if dto.UserID != tt.inquiries[i].UserID {
					t.Errorf("dto %d user id = %v, want %v", i, dto.UserID, tt.inquiries[i].UserID)
				}
			}
		})
	}
}

func TestInquiryUseCaseUpdateInquiry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     *domain.InquiryStatus
		repoErr    error
		wantErr    error
		wantStatus domain.InquiryStatus
	}{
		{name: "status nil", wantStatus: domain.InquiryPending},
		{name: "mark read", status: inquiryStatusPtr(domain.InquiryRead), wantStatus: domain.InquiryRead},
		{name: "mark replied", status: inquiryStatusPtr(domain.InquiryReplied), wantStatus: domain.InquiryReplied},
		{name: "close", status: inquiryStatusPtr(domain.InquiryClosed), wantStatus: domain.InquiryClosed},
		{name: "invalid status", status: inquiryStatusPtr(domain.InquiryStatus(99)), wantErr: domain.ErrInvalidInput},
		{name: "repo error", status: inquiryStatusPtr(domain.InquiryRead), repoErr: errFake, wantErr: errFake},
		{name: "not found", status: inquiryStatusPtr(domain.InquiryRead), repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inquiryRepo := newFakeInquiryRepo()
			if tt.repoErr != nil {
				inquiryRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
					return nil, tt.repoErr
				}
			}
			uc := usecases.NewInquiryUseCase(inquiryRepo, newFakeTimer())

			err := uc.UpdateInquiry(context.Background(), testInquiryID, dto.UpdateInquiryRequest{Status: tt.status})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if tt.wantErr == domain.ErrInvalidInput && len(inquiryRepo.updated) != 0 {
					t.Errorf("updated inquiries = %d, want 0", len(inquiryRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(inquiryRepo.updated) != 1 {
				t.Fatalf("updated inquiries = %d, want 1", len(inquiryRepo.updated))
			}
			if updated := inquiryRepo.updated[0]; updated.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", updated.Status, tt.wantStatus)
			}
		})
	}
}
