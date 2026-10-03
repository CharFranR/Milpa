package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewMatch(t *testing.T) {
	t.Parallel()

	offerID := uuid.New()
	requestID := uuid.New()

	match := NewMatch(offerID, requestID, 75.5, Kg)

	if match.ID == uuid.Nil {
		t.Error("expected a generated ID, got nil UUID")
	}
	if match.SupplyOffer != offerID {
		t.Errorf("supply_offer = %v, want %v", match.SupplyOffer, offerID)
	}
	if match.SupplyRequest != requestID {
		t.Errorf("supply_request = %v, want %v", match.SupplyRequest, requestID)
	}
	if match.MatchedAmount != 75.5 {
		t.Errorf("matched_amount = %v, want %v", match.MatchedAmount, 75.5)
	}
	if match.AmountUnit != Kg {
		t.Errorf("amount_unit = %v, want %v", match.AmountUnit, Kg)
	}
	if match.Status != MatchActive {
		t.Errorf("status = %v, want %v", match.Status, MatchActive)
	}
	if match.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if match.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestMatchCancel(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		status     MatchStatus
		wantErr    error
		wantStatus MatchStatus
	}{
		{name: "cancel active", status: MatchActive, wantErr: nil, wantStatus: MatchCancelled},
		{name: "cancel cancelled", status: MatchCancelled, wantErr: ErrInvalidMatchStatus, wantStatus: MatchCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match := &Match{Status: tt.status, UpdatedAt: now.Add(-time.Hour)}

			err := match.Cancel()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if match.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, match.Status)
				}
				if !match.UpdatedAt.Equal(now.Add(-time.Hour)) {
					t.Errorf("updated_at changed from %v on error", now.Add(-time.Hour))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if match.Status != MatchCancelled {
				t.Errorf("status = %v, want %v", match.Status, MatchCancelled)
			}
		})
	}
}

func TestMatchIsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status MatchStatus
		want   bool
	}{
		{name: "active", status: MatchActive, want: true},
		{name: "cancelled", status: MatchCancelled, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match := Match{Status: tt.status}

			if got := match.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status MatchStatus
		want   string
	}{
		{name: "active", status: MatchActive, want: "active"},
		{name: "cancelled", status: MatchCancelled, want: "cancelled"},
		{name: "unknown", status: MatchStatus(99), want: "unknown"},
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
