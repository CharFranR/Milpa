package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSupplyOffer(t *testing.T) {
	t.Parallel()

	supplierID := uuid.New()
	requestID := uuid.New()
	deliveryDay := time.Now().Add(48 * time.Hour)

	offer := NewSupplyOffer(supplierID, requestID, 120, Kg, deliveryDay, true)

	if offer.ID == uuid.Nil {
		t.Error("expected a generated ID, got nil UUID")
	}
	if offer.SupplierID != supplierID {
		t.Errorf("supplier_id = %v, want %v", offer.SupplierID, supplierID)
	}
	if offer.SupplyRequest != requestID {
		t.Errorf("supply_request = %v, want %v", offer.SupplyRequest, requestID)
	}
	if offer.TotalAmount != 120 {
		t.Errorf("total_amount = %v, want %v", offer.TotalAmount, 120)
	}
	if offer.AmountUnit != Kg {
		t.Errorf("amount_unit = %v, want %v", offer.AmountUnit, Kg)
	}
	if !offer.ProposedDeliveryDay.Equal(deliveryDay) {
		t.Errorf("proposed_delivery_day = %v, want %v", offer.ProposedDeliveryDay, deliveryDay)
	}
	if !offer.DeliveryAvailable {
		t.Error("delivery_available = false, want true")
	}
	if offer.Status != OfferActive {
		t.Errorf("status = %v, want %v", offer.Status, OfferActive)
	}
}

func TestSupplyOfferWithdraw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     OfferStatus
		wantErr    error
		wantStatus OfferStatus
	}{
		{name: "withdraw active", status: OfferActive, wantErr: nil, wantStatus: OfferWithdrawn},
		{name: "withdraw matched", status: OfferMatched, wantErr: ErrInvalidOfferStatus, wantStatus: OfferMatched},
		{name: "withdraw rejected", status: OfferRejected, wantErr: ErrInvalidOfferStatus, wantStatus: OfferRejected},
		{name: "withdraw withdrawn", status: OfferWithdrawn, wantErr: ErrInvalidOfferStatus, wantStatus: OfferWithdrawn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offer := &SupplyOffer{Status: tt.status}

			err := offer.Withdraw()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if offer.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, offer.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if offer.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", offer.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyOfferReject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     OfferStatus
		wantErr    error
		wantStatus OfferStatus
	}{
		{name: "reject active", status: OfferActive, wantErr: nil, wantStatus: OfferRejected},
		{name: "reject matched", status: OfferMatched, wantErr: ErrInvalidOfferStatus, wantStatus: OfferMatched},
		{name: "reject rejected", status: OfferRejected, wantErr: ErrInvalidOfferStatus, wantStatus: OfferRejected},
		{name: "reject withdrawn", status: OfferWithdrawn, wantErr: ErrInvalidOfferStatus, wantStatus: OfferWithdrawn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offer := &SupplyOffer{Status: tt.status}

			err := offer.Reject()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if offer.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, offer.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if offer.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", offer.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyOfferMarkMatched(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     OfferStatus
		wantErr    error
		wantStatus OfferStatus
	}{
		{name: "match active", status: OfferActive, wantErr: nil, wantStatus: OfferMatched},
		{name: "match already matched", status: OfferMatched, wantErr: ErrInvalidOfferStatus, wantStatus: OfferMatched},
		{name: "match rejected", status: OfferRejected, wantErr: ErrInvalidOfferStatus, wantStatus: OfferRejected},
		{name: "match withdrawn", status: OfferWithdrawn, wantErr: ErrInvalidOfferStatus, wantStatus: OfferWithdrawn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offer := &SupplyOffer{Status: tt.status}

			err := offer.MarkMatched()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if offer.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, offer.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if offer.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", offer.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyOfferIsActionable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status OfferStatus
		want   bool
	}{
		{name: "active", status: OfferActive, want: true},
		{name: "matched", status: OfferMatched, want: false},
		{name: "rejected", status: OfferRejected, want: false},
		{name: "withdrawn", status: OfferWithdrawn, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offer := SupplyOffer{Status: tt.status}

			if got := offer.IsActionable(); got != tt.want {
				t.Errorf("IsActionable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOfferStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status OfferStatus
		want   string
	}{
		{name: "active", status: OfferActive, want: "active"},
		{name: "matched", status: OfferMatched, want: "matched"},
		{name: "rejected", status: OfferRejected, want: "rejected"},
		{name: "withdrawn", status: OfferWithdrawn, want: "withdrawn"},
		{name: "unknown", status: OfferStatus(99), want: "unknown"},
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
