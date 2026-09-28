package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSupplyRequest(t *testing.T) {
	t.Parallel()

	buyerID := uuid.New()
	requestDeadline := time.Now().Add(24 * time.Hour)
	deliveryDeadline := time.Now().Add(72 * time.Hour)

	request := NewSupplyRequest(
		buyerID, "Maiz", 200, Kg, 10, 20, Lb,
		Address{}, requestDeadline, deliveryDeadline, "weekly restock", true,
	)

	if request.ID == uuid.Nil {
		t.Error("expected a generated ID, got nil UUID")
	}
	if request.BuyerID != buyerID {
		t.Errorf("buyer_id = %v, want %v", request.BuyerID, buyerID)
	}
	if request.ProductName != "Maiz" {
		t.Errorf("product_name = %q, want %q", request.ProductName, "Maiz")
	}
	if request.TotalAmount != 200 {
		t.Errorf("total_amount = %v, want %v", request.TotalAmount, 200)
	}
	if request.ActualAmount != 200 {
		t.Errorf("actual_amount = %v, want %v", request.ActualAmount, 200)
	}
	if !request.MultipleProviders {
		t.Error("multiple_providers = false, want true")
	}
	if request.Status != SupplyRequestOpen {
		t.Errorf("status = %v, want %v", request.Status, SupplyRequestOpen)
	}
	if !request.IsOpen() {
		t.Error("IsOpen() = false on a new request")
	}
	if request.RemainingAmount() != 200 {
		t.Errorf("RemainingAmount() = %v, want %v", request.RemainingAmount(), 200)
	}
}

func TestSupplyRequestCancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     SupplyRequestStatus
		wantErr    error
		wantStatus SupplyRequestStatus
	}{
		{name: "cancel open", status: SupplyRequestOpen, wantErr: nil, wantStatus: SupplyRequestCancelled},
		{name: "cancel cancelled", status: SupplyRequestCancelled, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCancelled},
		{name: "cancel completed", status: SupplyRequestCompleted, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCompleted},
		{name: "cancel expired", status: SupplyRequestExpired, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := &SupplyRequest{Status: tt.status}

			err := request.Cancel()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if request.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, request.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if request.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", request.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyRequestComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     SupplyRequestStatus
		wantErr    error
		wantStatus SupplyRequestStatus
	}{
		{name: "complete open", status: SupplyRequestOpen, wantErr: nil, wantStatus: SupplyRequestCompleted},
		{name: "complete cancelled", status: SupplyRequestCancelled, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCancelled},
		{name: "complete completed", status: SupplyRequestCompleted, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCompleted},
		{name: "complete expired", status: SupplyRequestExpired, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := &SupplyRequest{Status: tt.status}

			err := request.Complete()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if request.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, request.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if request.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", request.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyRequestExpire(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     SupplyRequestStatus
		wantErr    error
		wantStatus SupplyRequestStatus
	}{
		{name: "expire open", status: SupplyRequestOpen, wantErr: nil, wantStatus: SupplyRequestExpired},
		{name: "expire cancelled", status: SupplyRequestCancelled, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCancelled},
		{name: "expire completed", status: SupplyRequestCompleted, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestCompleted},
		{name: "expire expired", status: SupplyRequestExpired, wantErr: ErrInvalidRequestStatus, wantStatus: SupplyRequestExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := &SupplyRequest{Status: tt.status}

			err := request.Expire()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if request.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, request.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if request.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", request.Status, tt.wantStatus)
			}
		})
	}
}

func TestSupplyRequestReserveAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		status       SupplyRequestStatus
		totalAmount  float64
		actualAmount float64
		amount       float64
		wantErr      error
		wantActual   float64
	}{
		{
			name:         "reserve partial from open",
			status:       SupplyRequestOpen,
			totalAmount:  200,
			actualAmount: 200,
			amount:       40,
			wantErr:      nil,
			wantActual:   160,
		},
		{
			name:         "reserve more than remaining",
			status:       SupplyRequestOpen,
			totalAmount:  200,
			actualAmount: 100,
			amount:       150,
			wantErr:      ErrInsufficientAmount,
			wantActual:   100,
		},
		{
			name:         "reserve negative amount",
			status:       SupplyRequestOpen,
			totalAmount:  200,
			actualAmount: 200,
			amount:       -10,
			wantErr:      ErrInvalidInput,
			wantActual:   200,
		},
		{
			name:         "reserve zero amount",
			status:       SupplyRequestOpen,
			totalAmount:  200,
			actualAmount: 200,
			amount:       0,
			wantErr:      ErrInvalidInput,
			wantActual:   200,
		},
		{
			name:         "reserve after cancel",
			status:       SupplyRequestCancelled,
			totalAmount:  200,
			actualAmount: 200,
			amount:       10,
			wantErr:      ErrInvalidRequestStatus,
			wantActual:   200,
		},
		{
			name:         "reserve after completion",
			status:       SupplyRequestCompleted,
			totalAmount:  200,
			actualAmount: 200,
			amount:       10,
			wantErr:      ErrInvalidRequestStatus,
			wantActual:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := &SupplyRequest{
				TotalAmount:  tt.totalAmount,
				ActualAmount: tt.actualAmount,
				Status:       tt.status,
			}

			err := request.ReserveAmount(tt.amount)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if request.ActualAmount != tt.wantActual {
				t.Errorf("actual_amount = %v, want %v", request.ActualAmount, tt.wantActual)
			}
		})
	}
}

func TestSupplyRequestReleaseAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		totalAmount  float64
		actualAmount float64
		amount       float64
		wantErr      error
		wantActual   float64
	}{
		{
			name:         "release adds back",
			totalAmount:  200,
			actualAmount: 160,
			amount:       30,
			wantErr:      nil,
			wantActual:   190,
		},
		{
			name:         "release caps at total",
			totalAmount:  200,
			actualAmount: 160,
			amount:       100,
			wantErr:      nil,
			wantActual:   200,
		},
		{
			name:         "release never goes negative",
			totalAmount:  200,
			actualAmount: -5,
			amount:       1,
			wantErr:      nil,
			wantActual:   0,
		},
		{
			name:         "release negative amount",
			totalAmount:  200,
			actualAmount: 160,
			amount:       -10,
			wantErr:      ErrInvalidInput,
			wantActual:   160,
		},
		{
			name:         "release zero amount",
			totalAmount:  200,
			actualAmount: 160,
			amount:       0,
			wantErr:      ErrInvalidInput,
			wantActual:   160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := &SupplyRequest{
				TotalAmount:  tt.totalAmount,
				ActualAmount: tt.actualAmount,
				Status:       SupplyRequestOpen,
			}

			err := request.ReleaseAmount(tt.amount)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if request.ActualAmount != tt.wantActual {
				t.Errorf("actual_amount = %v, want %v", request.ActualAmount, tt.wantActual)
			}
		})
	}
}

func TestSupplyRequestIsOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status SupplyRequestStatus
		want   bool
	}{
		{name: "open", status: SupplyRequestOpen, want: true},
		{name: "cancelled", status: SupplyRequestCancelled, want: false},
		{name: "completed", status: SupplyRequestCompleted, want: false},
		{name: "expired", status: SupplyRequestExpired, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := SupplyRequest{Status: tt.status}

			if got := request.IsOpen(); got != tt.want {
				t.Errorf("IsOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupplyRequestStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status SupplyRequestStatus
		want   string
	}{
		{name: "open", status: SupplyRequestOpen, want: "open"},
		{name: "cancelled", status: SupplyRequestCancelled, want: "cancelled"},
		{name: "completed", status: SupplyRequestCompleted, want: "completed"},
		{name: "expired", status: SupplyRequestExpired, want: "expired"},
		{name: "unknown", status: SupplyRequestStatus(99), want: "unknown"},
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
