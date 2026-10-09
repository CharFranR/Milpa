package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUnitOfMeasure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		unitName string
		wantCode string
		wantErr  error
	}{
		{name: "happy path normalizes the code", code: "  KG  ", unitName: "Kilogramo", wantCode: "kg"},
		{name: "empty name", code: "kg", unitName: "", wantErr: ErrNameRequired},
		{name: "empty code", code: "   ", unitName: "Kilogramo", wantErr: ErrCodeRequired},
		{
			name:     "21 character code",
			code:     strings.Repeat("a", 21),
			unitName: "Kilogramo",
			wantErr:  ErrInvalidInput,
		},
		{
			name:     "61 character name",
			code:     "kg",
			unitName: strings.Repeat("a", 61),
			wantErr:  ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			unit, err := NewUnitOfMeasure(tt.code, tt.unitName)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewUnitOfMeasure() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUnitOfMeasure() error: %v", err)
			}
			if unit == nil {
				t.Fatal("NewUnitOfMeasure() returned a nil unit")
			}
			if unit.ID.String() == "" || unit.ID == [16]byte{} {
				t.Error("NewUnitOfMeasure() left the id unset")
			}
			if unit.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", unit.Code, tt.wantCode)
			}
			if unit.Name != tt.unitName {
				t.Errorf("name = %q, want %q", unit.Name, tt.unitName)
			}
			if !unit.IsActive {
				t.Error("NewUnitOfMeasure() stored the unit inactive")
			}
		})
	}
}

func TestUnitOfMeasureSetCode(t *testing.T) {
	t.Parallel()

	unit := &UnitOfMeasure{}

	if err := unit.SetCode("  Qq  "); err != nil {
		t.Fatalf("SetCode() error: %v", err)
	}
	if unit.Code != "qq" {
		t.Errorf("code = %q, want the trimmed lowercase qq", unit.Code)
	}

	if err := unit.SetCode("   "); !errors.Is(err, ErrCodeRequired) {
		t.Errorf("SetCode(blank) error = %v, want ErrCodeRequired", err)
	}
	if unit.Code != "qq" {
		t.Errorf("a refused SetCode changed the code to %q", unit.Code)
	}

	if err := unit.SetCode(strings.Repeat("a", 21)); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("SetCode(21 chars) error = %v, want ErrInvalidInput", err)
	}
	if unit.Code != "qq" {
		t.Errorf("an over-long SetCode changed the code to %q", unit.Code)
	}
}

func TestUnitOfMeasureSetName(t *testing.T) {
	t.Parallel()

	unit := &UnitOfMeasure{}

	if err := unit.SetName("  Kilogramo  "); err != nil {
		t.Fatalf("SetName() error: %v", err)
	}
	if unit.Name != "Kilogramo" {
		t.Errorf("name = %q, want the trimmed Kilogramo", unit.Name)
	}

	if err := unit.SetName("   "); !errors.Is(err, ErrNameRequired) {
		t.Errorf("SetName(blank) error = %v, want ErrNameRequired", err)
	}
	if unit.Name != "Kilogramo" {
		t.Errorf("a refused SetName changed the name to %q", unit.Name)
	}

	if err := unit.SetName(strings.Repeat("a", 61)); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("SetName(61 chars) error = %v, want ErrInvalidInput", err)
	}
	if unit.Name != "Kilogramo" {
		t.Errorf("an over-long SetName changed the name to %q", unit.Name)
	}
}

func TestUnitOfMeasureActivateDeactivate(t *testing.T) {
	t.Parallel()

	unit := &UnitOfMeasure{IsActive: true}

	unit.Deactivate()
	if unit.IsActive {
		t.Error("Deactivate() left the unit active")
	}

	unit.Activate()
	if !unit.IsActive {
		t.Error("Activate() left the unit inactive")
	}
}
