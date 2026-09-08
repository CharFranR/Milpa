package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewInquiry(t *testing.T) {
	t.Parallel()

	now := time.Now()
	userID := uuid.New()
	offeringID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		offeringID uuid.UUID
		message    string
		now        time.Time
		wantErr    error
	}{
		{name: "happy path", userID: userID, offeringID: offeringID, message: "How much?", now: now},
		{name: "empty message", userID: userID, offeringID: offeringID, now: now, wantErr: ErrMessageRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inquiry, err := NewInquiry(tt.userID, tt.offeringID, tt.message, tt.now)

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
			if inquiry.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if inquiry.UserID != tt.userID {
				t.Errorf("user id = %v, want %v", inquiry.UserID, tt.userID)
			}
			if inquiry.OfferingID != tt.offeringID {
				t.Errorf("offering id = %v, want %v", inquiry.OfferingID, tt.offeringID)
			}
			if inquiry.Message != tt.message {
				t.Errorf("message = %q, want %q", inquiry.Message, tt.message)
			}
			if inquiry.Status != InquiryPending {
				t.Errorf("status = %v, want %v", inquiry.Status, InquiryPending)
			}
			if !inquiry.CreatedAt.Equal(tt.now) {
				t.Errorf("created at = %v, want %v", inquiry.CreatedAt, tt.now)
			}
		})
	}
}

func TestInquiryMarkAsRead(t *testing.T) {
	t.Parallel()

	inquiry := &Inquiry{Status: InquiryPending}

	inquiry.MarkAsRead()

	if inquiry.Status != InquiryRead {
		t.Errorf("Status = %v, want %v", inquiry.Status, InquiryRead)
	}
}

func TestInquiryMarkAsReplied(t *testing.T) {
	t.Parallel()

	inquiry := &Inquiry{Status: InquiryPending}

	inquiry.MarkAsReplied()

	if inquiry.Status != InquiryReplied {
		t.Errorf("Status = %v, want %v", inquiry.Status, InquiryReplied)
	}
}

func TestInquiryClose(t *testing.T) {
	t.Parallel()

	inquiry := &Inquiry{Status: InquiryPending}

	inquiry.Close()

	if inquiry.Status != InquiryClosed {
		t.Errorf("Status = %v, want %v", inquiry.Status, InquiryClosed)
	}
}

func TestInquiryStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status InquiryStatus
		want   string
	}{
		{name: "pending", status: InquiryPending, want: "pending"},
		{name: "read", status: InquiryRead, want: "read"},
		{name: "replied", status: InquiryReplied, want: "replied"},
		{name: "closed", status: InquiryClosed, want: "closed"},
		{name: "unknown value", status: InquiryStatus(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
