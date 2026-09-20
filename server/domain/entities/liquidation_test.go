package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewLiquidation(t *testing.T) {
	t.Parallel()

	now := time.Now()
	supplierID := uuid.New()

	tests := []struct {
		name          string
		supplierID    uuid.UUID
		productName   string
		quantity      float64
		unitOfMeasure string
		totalPrice    float64
		unitPrice     float64
		now           time.Time
		wantErr       error
	}{
		{
			name:          "happy path",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      100,
			unitOfMeasure: "kg",
			totalPrice:    5000,
			unitPrice:     50,
			now:           now,
		},
		{
			name:          "empty product name",
			supplierID:    supplierID,
			productName:   "",
			quantity:      100,
			unitOfMeasure: "kg",
			totalPrice:    5000,
			unitPrice:     50,
			now:           now,
			wantErr:       ErrProductNameRequired,
		},
		{
			name:          "zero quantity",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      0,
			unitOfMeasure: "kg",
			totalPrice:    5000,
			unitPrice:     50,
			now:           now,
			wantErr:       ErrInvalidQuantity,
		},
		{
			name:          "negative quantity",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      -10,
			unitOfMeasure: "kg",
			totalPrice:    5000,
			unitPrice:     50,
			now:           now,
			wantErr:       ErrInvalidQuantity,
		},
		{
			name:          "empty unit of measure",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      100,
			unitOfMeasure: "",
			totalPrice:    5000,
			unitPrice:     50,
			now:           now,
			wantErr:       ErrUnitOfMeasureRequired,
		},
		{
			name:          "zero total price",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      100,
			unitOfMeasure: "kg",
			totalPrice:    0,
			unitPrice:     50,
			now:           now,
			wantErr:       ErrInvalidPrice,
		},
		{
			name:          "zero unit price",
			supplierID:    supplierID,
			productName:   "Maiz",
			quantity:      100,
			unitOfMeasure: "kg",
			totalPrice:    5000,
			unitPrice:     0,
			now:           now,
			wantErr:       ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq, err := NewLiquidation(
				tt.supplierID,
				tt.productName,
				tt.quantity,
				tt.unitOfMeasure,
				tt.totalPrice,
				tt.unitPrice,
				tt.now,
			)

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
			if liq.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if liq.SupplierID != tt.supplierID {
				t.Errorf("supplier_id = %v, want %v", liq.SupplierID, tt.supplierID)
			}
			if liq.ProductName != tt.productName {
				t.Errorf("product_name = %q, want %q", liq.ProductName, tt.productName)
			}
			if liq.Quantity != tt.quantity {
				t.Errorf("quantity = %v, want %v", liq.Quantity, tt.quantity)
			}
			if liq.UnitOfMeasure != tt.unitOfMeasure {
				t.Errorf("unit_of_measure = %q, want %q", liq.UnitOfMeasure, tt.unitOfMeasure)
			}
			if liq.TotalPrice != tt.totalPrice {
				t.Errorf("total_price = %v, want %v", liq.TotalPrice, tt.totalPrice)
			}
			if liq.UnitPrice != tt.unitPrice {
				t.Errorf("unit_price = %v, want %v", liq.UnitPrice, tt.unitPrice)
			}
			if liq.Status != LiquidationOpen {
				t.Errorf("status = %v, want %v", liq.Status, LiquidationOpen)
			}
			if liq.Visibility != "public" {
				t.Errorf("visibility = %q, want %q", liq.Visibility, "public")
			}
			if liq.AllocationMethod != AllocationManual {
				t.Errorf("allocation_method = %v, want %v", liq.AllocationMethod, AllocationManual)
			}
			if !liq.CreatedAt.Equal(tt.now) {
				t.Errorf("created_at = %v, want %v", liq.CreatedAt, tt.now)
			}
			if !liq.UpdatedAt.Equal(tt.now) {
				t.Errorf("updated_at = %v, want %v", liq.UpdatedAt, tt.now)
			}
		})
	}
}

func TestLiquidationIsOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status LiquidationStatus
		want   bool
	}{
		{name: "open", status: LiquidationOpen, want: true},
		{name: "closed", status: LiquidationClosed, want: false},
		{name: "expired", status: LiquidationExpired, want: false},
		{name: "assigned", status: LiquidationAssigned, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := Liquidation{Status: tt.status}

			if got := liq.IsOpen(); got != tt.want {
				t.Errorf("IsOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidationCanReceiveInterest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status LiquidationStatus
		want   bool
	}{
		{name: "open", status: LiquidationOpen, want: true},
		{name: "closed", status: LiquidationClosed, want: false},
		{name: "expired", status: LiquidationExpired, want: false},
		{name: "assigned", status: LiquidationAssigned, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := Liquidation{Status: tt.status}

			if got := liq.CanReceiveInterest(); got != tt.want {
				t.Errorf("CanReceiveInterest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidationIsExpired(t *testing.T) {
	t.Parallel()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		now       time.Time
		want      bool
	}{
		{name: "no expiry", expiresAt: nil, now: now, want: false},
		{name: "not expired yet", expiresAt: &future, now: now, want: false},
		{name: "already expired", expiresAt: &past, now: now, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := Liquidation{ExpiresAt: tt.expiresAt}

			if got := liq.IsExpired(tt.now); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidationClose(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name    string
		status  LiquidationStatus
		wantErr error
	}{
		{name: "close open", status: LiquidationOpen, wantErr: nil},
		{name: "close closed", status: LiquidationClosed, wantErr: ErrLiquidationNotOpen},
		{name: "close expired", status: LiquidationExpired, wantErr: ErrLiquidationNotOpen},
		{name: "close assigned", status: LiquidationAssigned, wantErr: ErrLiquidationNotOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := &Liquidation{Status: tt.status, UpdatedAt: earlier}

			err := liq.Close(now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if liq.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, liq.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if liq.Status != LiquidationClosed {
				t.Errorf("status = %v, want %v", liq.Status, LiquidationClosed)
			}
			if liq.ClosedAt == nil || !liq.ClosedAt.Equal(now) {
				t.Errorf("closed_at = %v, want %v", liq.ClosedAt, now)
			}
			if !liq.UpdatedAt.Equal(now) {
				t.Errorf("updated_at = %v, want %v", liq.UpdatedAt, now)
			}
		})
	}
}

func TestLiquidationExpire(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name    string
		status  LiquidationStatus
		wantErr error
	}{
		{name: "expire open", status: LiquidationOpen, wantErr: nil},
		{name: "expire closed", status: LiquidationClosed, wantErr: ErrLiquidationNotOpen},
		{name: "expire expired", status: LiquidationExpired, wantErr: ErrLiquidationNotOpen},
		{name: "expire assigned", status: LiquidationAssigned, wantErr: ErrLiquidationNotOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := &Liquidation{Status: tt.status, UpdatedAt: earlier}

			err := liq.Expire(now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if liq.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, liq.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if liq.Status != LiquidationExpired {
				t.Errorf("status = %v, want %v", liq.Status, LiquidationExpired)
			}
			if liq.ClosedAt == nil || !liq.ClosedAt.Equal(now) {
				t.Errorf("closed_at = %v, want %v", liq.ClosedAt, now)
			}
			if !liq.UpdatedAt.Equal(now) {
				t.Errorf("updated_at = %v, want %v", liq.UpdatedAt, now)
			}
		})
	}
}

func TestLiquidationAssign(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name    string
		status  LiquidationStatus
		wantErr error
	}{
		{name: "assign open", status: LiquidationOpen, wantErr: nil},
		{name: "assign closed", status: LiquidationClosed, wantErr: ErrLiquidationCannotAssign},
		{name: "assign expired", status: LiquidationExpired, wantErr: ErrLiquidationCannotAssign},
		{name: "assign assigned", status: LiquidationAssigned, wantErr: ErrLiquidationCannotAssign},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := &Liquidation{Status: tt.status, UpdatedAt: earlier}

			err := liq.Assign(now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if liq.Status != tt.status {
					t.Errorf("status changed from %v to %v on error", tt.status, liq.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if liq.Status != LiquidationAssigned {
				t.Errorf("status = %v, want %v", liq.Status, LiquidationAssigned)
			}
			if !liq.UpdatedAt.Equal(now) {
				t.Errorf("updated_at = %v, want %v", liq.UpdatedAt, now)
			}
		})
	}
}

func TestLiquidationUpdateVisibility(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name       string
		visibility string
		wantErr    error
	}{
		{name: "set public", visibility: "public", wantErr: nil},
		{name: "set private", visibility: "private", wantErr: nil},
		{name: "invalid visibility", visibility: "invalid", wantErr: ErrInvalidVisibility},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			liq := &Liquidation{Visibility: "public", UpdatedAt: earlier}

			err := liq.UpdateVisibility(tt.visibility, now)

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
			if liq.Visibility != tt.visibility {
				t.Errorf("visibility = %q, want %q", liq.Visibility, tt.visibility)
			}
			if !liq.UpdatedAt.Equal(now) {
				t.Errorf("updated_at = %v, want %v", liq.UpdatedAt, now)
			}
		})
	}
}

func TestLiquidationStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status LiquidationStatus
		want   string
	}{
		{name: "open", status: LiquidationOpen, want: "open"},
		{name: "closed", status: LiquidationClosed, want: "closed"},
		{name: "expired", status: LiquidationExpired, want: "expired"},
		{name: "assigned", status: LiquidationAssigned, want: "assigned"},
		{name: "unknown", status: LiquidationStatus(99), want: "unknown"},
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

func TestAllocationMethodString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method AllocationMethod
		want   string
	}{
		{name: "manual", method: AllocationManual, want: "manual"},
		{name: "unknown", method: AllocationMethod(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.method.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
