package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewSupplierInventory(t *testing.T) {
	t.Parallel()

	supplierID := uuid.New()

	inventory := NewSupplierInventory(supplierID, "Maiz", 500, Kg)

	if inventory.ID == uuid.Nil {
		t.Error("expected a generated ID, got nil UUID")
	}
	if inventory.SupplierID != supplierID {
		t.Errorf("supplier_id = %v, want %v", inventory.SupplierID, supplierID)
	}
	if inventory.ProductName != "Maiz" {
		t.Errorf("product_name = %q, want %q", inventory.ProductName, "Maiz")
	}
	if inventory.Quantity != 500 {
		t.Errorf("quantity = %v, want %v", inventory.Quantity, 500)
	}
	if inventory.AmountUnit != Kg {
		t.Errorf("amount_unit = %v, want %v", inventory.AmountUnit, Kg)
	}
	if inventory.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if inventory.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestSupplierInventorySetQuantity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		quantity    float64
		wantErr     error
		wantOutcome float64
	}{
		{name: "set positive quantity", quantity: 250, wantErr: nil, wantOutcome: 250},
		{name: "set zero quantity", quantity: 0, wantErr: nil, wantOutcome: 0},
		{name: "set negative quantity", quantity: -10, wantErr: ErrInvalidQuantity, wantOutcome: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inventory := NewSupplierInventory(uuid.New(), "Maiz", 100, Kg)

			err := inventory.SetQuantity(tt.quantity)

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

			if inventory.Quantity != tt.wantOutcome {
				t.Errorf("quantity = %v, want %v", inventory.Quantity, tt.wantOutcome)
			}
		})
	}
}
