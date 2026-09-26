package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransaction(t *testing.T) {
	t.Parallel()

	matchID := uuid.New()

	trx := NewTransaction(matchID)

	if trx.ID == uuid.Nil {
		t.Error("expected a generated ID, got nil UUID")
	}
	if trx.MatchID != matchID {
		t.Errorf("match_id = %v, want %v", trx.MatchID, matchID)
	}
	if trx.Status != TransactionMatched {
		t.Errorf("status = %v, want %v", trx.Status, TransactionMatched)
	}
	if trx.BuyerStartConfirmedAt != nil {
		t.Errorf("buyer_start_confirmed_at = %v, want nil", trx.BuyerStartConfirmedAt)
	}
	if trx.SupplierStartConfirmedAt != nil {
		t.Errorf("supplier_start_confirmed_at = %v, want nil", trx.SupplierStartConfirmedAt)
	}
	if trx.BuyerDeliveryConfirmedAt != nil {
		t.Errorf("buyer_delivery_confirmed_at = %v, want nil", trx.BuyerDeliveryConfirmedAt)
	}
	if trx.SupplierDeliveryConfirmedAt != nil {
		t.Errorf("supplier_delivery_confirmed_at = %v, want nil", trx.SupplierDeliveryConfirmedAt)
	}
	if trx.CancelledBy != nil {
		t.Errorf("cancelled_by = %v, want nil", trx.CancelledBy)
	}
	if trx.CancelReason != "" {
		t.Errorf("cancel_reason = %q, want empty", trx.CancelReason)
	}
	if trx.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if trx.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestTransactionConfirmStart(t *testing.T) {
	t.Parallel()

	now := time.Now()
	reason := "buyer changed mind"

	matched := func() *Transaction {
		return NewTransaction(uuid.New())
	}
	buyerPreConfirmed := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.ConfirmStart(BuyerParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	supplierPreConfirmed := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.ConfirmStart(SupplierParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	inProgress := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.ConfirmStart(BuyerParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		if err := trx.ConfirmStart(SupplierParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	completed := func() *Transaction {
		trx := inProgress()
		if err := trx.ConfirmDelivery(BuyerParticipant, now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		if err := trx.ConfirmDelivery(SupplierParticipant, now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	cancelled := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.Cancel(uuid.New(), reason, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-cancel error: %v", err)
		}
		return trx
	}

	tests := []struct {
		name        string
		build       func() *Transaction
		participant TransactionParticipant
		wantErr     error
		wantStatus  TransactionStatus
	}{
		{name: "confirm buyer from matched", build: matched, participant: BuyerParticipant, wantErr: nil, wantStatus: TransactionMatched},
		{name: "confirm supplier after buyer", build: buyerPreConfirmed, participant: SupplierParticipant, wantErr: nil, wantStatus: TransactionInProgress},
		{name: "confirm buyer after supplier", build: supplierPreConfirmed, participant: BuyerParticipant, wantErr: nil, wantStatus: TransactionInProgress},
		{name: "buyer already confirmed", build: buyerPreConfirmed, participant: BuyerParticipant, wantErr: ErrAlreadyConfirmed, wantStatus: TransactionMatched},
		{name: "supplier already confirmed", build: supplierPreConfirmed, participant: SupplierParticipant, wantErr: ErrAlreadyConfirmed, wantStatus: TransactionMatched},
		{name: "confirm from in progress", build: inProgress, participant: BuyerParticipant, wantErr: ErrInvalidTransactionTransition, wantStatus: TransactionInProgress},
		{name: "confirm from completed", build: completed, participant: BuyerParticipant, wantErr: ErrInvalidTransactionTransition, wantStatus: TransactionCompleted},
		{name: "confirm from cancelled", build: cancelled, participant: BuyerParticipant, wantErr: ErrInvalidTransactionTransition, wantStatus: TransactionCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			trx := tt.build()

			err := trx.ConfirmStart(tt.participant, now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if trx.Status != tt.wantStatus {
					t.Errorf("status changed from %v to %v on error", tt.wantStatus, trx.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if trx.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", trx.Status, tt.wantStatus)
			}
		})
	}
}

func TestTransactionConfirmStartRecordsTimestamps(t *testing.T) {
	t.Parallel()

	now := time.Now()
	trx := NewTransaction(uuid.New())

	if err := trx.ConfirmStart(BuyerParticipant, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trx.BuyerStartConfirmedAt == nil || !trx.BuyerStartConfirmedAt.Equal(now) {
		t.Errorf("buyer_start_confirmed_at = %v, want %v", trx.BuyerStartConfirmedAt, now)
	}
	if trx.SupplierStartConfirmedAt != nil {
		t.Errorf("supplier_start_confirmed_at = %v, want nil", trx.SupplierStartConfirmedAt)
	}
	if !trx.UpdatedAt.Equal(now) {
		t.Errorf("updated_at = %v, want %v", trx.UpdatedAt, now)
	}
}

func TestTransactionConfirmDelivery(t *testing.T) {
	t.Parallel()

	now := time.Now()

	inProgress := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.ConfirmStart(BuyerParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		if err := trx.ConfirmStart(SupplierParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	buyerDelivered := func() *Transaction {
		trx := inProgress()
		if err := trx.ConfirmDelivery(BuyerParticipant, now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	cancelled := func() *Transaction {
		trx := inProgress()
		if err := trx.Cancel(uuid.New(), "cancelled", now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-cancel error: %v", err)
		}
		return trx
	}
	matched := func() *Transaction {
		return NewTransaction(uuid.New())
	}

	tests := []struct {
		name        string
		build       func() *Transaction
		participant TransactionParticipant
		wantErr     error
		wantStatus  TransactionStatus
	}{
		{name: "deliver buyer from in progress", build: inProgress, participant: BuyerParticipant, wantErr: nil, wantStatus: TransactionInProgress},
		{name: "deliver supplier after buyer", build: buyerDelivered, participant: SupplierParticipant, wantErr: nil, wantStatus: TransactionCompleted},
		{name: "buyer already delivered", build: buyerDelivered, participant: BuyerParticipant, wantErr: ErrAlreadyConfirmed, wantStatus: TransactionInProgress},
		{name: "deliver from matched", build: matched, participant: BuyerParticipant, wantErr: ErrInvalidTransactionTransition, wantStatus: TransactionMatched},
		{name: "deliver from cancelled", build: cancelled, participant: BuyerParticipant, wantErr: ErrInvalidTransactionTransition, wantStatus: TransactionCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			trx := tt.build()

			err := trx.ConfirmDelivery(tt.participant, now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if trx.Status != tt.wantStatus {
					t.Errorf("status changed from %v to %v on error", tt.wantStatus, trx.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if trx.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", trx.Status, tt.wantStatus)
			}
		})
	}
}

func TestTransactionConfirmDeliveryRecordsTimestamps(t *testing.T) {
	t.Parallel()

	now := time.Now()
	trx := NewTransaction(uuid.New())

	if err := trx.ConfirmStart(BuyerParticipant, now.Add(-time.Hour)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := trx.ConfirmStart(SupplierParticipant, now.Add(-time.Hour)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := trx.ConfirmDelivery(BuyerParticipant, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trx.BuyerDeliveryConfirmedAt == nil || !trx.BuyerDeliveryConfirmedAt.Equal(now) {
		t.Errorf("buyer_delivery_confirmed_at = %v, want %v", trx.BuyerDeliveryConfirmedAt, now)
	}
	if trx.SupplierDeliveryConfirmedAt != nil {
		t.Errorf("supplier_delivery_confirmed_at = %v, want nil", trx.SupplierDeliveryConfirmedAt)
	}
}

func TestTransactionFullLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	trx := NewTransaction(uuid.New())

	if err := trx.ConfirmStart(BuyerParticipant, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := trx.ConfirmStart(SupplierParticipant, now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trx.Status != TransactionInProgress {
		t.Fatalf("status = %v, want %v", trx.Status, TransactionInProgress)
	}
	if err := trx.ConfirmDelivery(BuyerParticipant, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := trx.ConfirmDelivery(SupplierParticipant, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trx.Status != TransactionCompleted {
		t.Fatalf("status = %v, want %v", trx.Status, TransactionCompleted)
	}
	if trx.IsActive() {
		t.Error("IsActive() = true on completed transaction")
	}
}

func TestTransactionCancel(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cancelledBy := uuid.New()

	matched := func() *Transaction {
		return NewTransaction(uuid.New())
	}
	inProgress := func() *Transaction {
		trx := NewTransaction(uuid.New())
		if err := trx.ConfirmStart(BuyerParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		if err := trx.ConfirmStart(SupplierParticipant, now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	completed := func() *Transaction {
		trx := inProgress()
		if err := trx.ConfirmDelivery(BuyerParticipant, now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		if err := trx.ConfirmDelivery(SupplierParticipant, now.Add(-time.Minute)); err != nil {
			t.Fatalf("unexpected pre-confirm error: %v", err)
		}
		return trx
	}
	cancelled := func() *Transaction {
		trx := matched()
		if err := trx.Cancel(uuid.New(), "already cancelled", now.Add(-time.Hour)); err != nil {
			t.Fatalf("unexpected pre-cancel error: %v", err)
		}
		return trx
	}

	tests := []struct {
		name    string
		build   func() *Transaction
		reason  string
		wantErr error
	}{
		{name: "cancel from matched", build: matched, reason: "buyer changed mind", wantErr: nil},
		{name: "cancel from in progress", build: inProgress, reason: "delivery dispute", wantErr: nil},
		{name: "cancel from completed", build: completed, reason: "too late", wantErr: ErrTerminalState},
		{name: "cancel from cancelled", build: cancelled, reason: "again", wantErr: ErrTerminalState},
		{name: "empty reason from matched", build: matched, reason: "", wantErr: ErrReasonRequired},
		{name: "empty reason from in progress", build: inProgress, reason: "", wantErr: ErrReasonRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			trx := tt.build()
			previousStatus := trx.Status
			previousCancelledBy := trx.CancelledBy

			err := trx.Cancel(cancelledBy, tt.reason, now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if trx.Status != previousStatus {
					t.Errorf("status changed from %v to %v on error", previousStatus, trx.Status)
				}
				if trx.CancelledBy != previousCancelledBy {
					t.Errorf("cancelled_by changed from %v to %v on error", previousCancelledBy, trx.CancelledBy)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if trx.Status != TransactionCancelled {
				t.Errorf("status = %v, want %v", trx.Status, TransactionCancelled)
			}
			if trx.CancelledBy == nil || *trx.CancelledBy != cancelledBy {
				t.Errorf("cancelled_by = %v, want %v", trx.CancelledBy, cancelledBy)
			}
			if trx.CancelReason != tt.reason {
				t.Errorf("cancel_reason = %q, want %q", trx.CancelReason, tt.reason)
			}
			if !trx.UpdatedAt.Equal(now) {
				t.Errorf("updated_at = %v, want %v", trx.UpdatedAt, now)
			}
		})
	}
}

func TestTransactionIsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status TransactionStatus
		want   bool
	}{
		{name: "matched", status: TransactionMatched, want: true},
		{name: "in progress", status: TransactionInProgress, want: true},
		{name: "completed", status: TransactionCompleted, want: false},
		{name: "cancelled", status: TransactionCancelled, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			trx := Transaction{Status: tt.status}

			if got := trx.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status TransactionStatus
		want   string
	}{
		{name: "matched", status: TransactionMatched, want: "matched"},
		{name: "in progress", status: TransactionInProgress, want: "in_progress"},
		{name: "completed", status: TransactionCompleted, want: "completed"},
		{name: "cancelled", status: TransactionCancelled, want: "cancelled"},
		{name: "unknown", status: TransactionStatus(99), want: "unknown"},
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
