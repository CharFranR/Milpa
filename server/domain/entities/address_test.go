package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		department   string
		municipality string
		addressLine  string
		wantErr      error
	}{
		{name: "happy path", department: "Managua", municipality: "Managua", addressLine: "123 Main St"},
		{name: "empty department", municipality: "Managua", addressLine: "123 Main St", wantErr: ErrDepartmentRequired},
		{name: "empty municipality", department: "Managua", addressLine: "123 Main St", wantErr: ErrMunicipalityRequired},
		{name: "empty address line", department: "Managua", municipality: "Managua", wantErr: ErrAddressLineRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			address, err := NewAddress(tt.department, tt.municipality, tt.addressLine)

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
			if address.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if address.Department != tt.department {
				t.Errorf("department = %q, want %q", address.Department, tt.department)
			}
			if address.Municipality != tt.municipality {
				t.Errorf("municipality = %q, want %q", address.Municipality, tt.municipality)
			}
			if address.AddressLine != tt.addressLine {
				t.Errorf("address line = %q, want %q", address.AddressLine, tt.addressLine)
			}
			if address.Latitude != 0 {
				t.Errorf("latitude = %v, want zero value", address.Latitude)
			}
			if address.Longitude != 0 {
				t.Errorf("longitude = %v, want zero value", address.Longitude)
			}
		})
	}
}

func TestAddressFullAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		address Address
		want    string
	}{
		{name: "all fields set", address: Address{AddressLine: "123 Main St", Municipality: "Managua", Department: "Managua"}, want: "123 Main St, Managua, Managua"},
		{name: "zero value", want: ", , "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.address.FullAddress(); got != tt.want {
				t.Errorf("FullAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressHasCoordinates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		want      bool
	}{
		{name: "no coordinates", want: false},
		{name: "latitude only", latitude: 12.114, want: false},
		{name: "longitude only", longitude: -86.236, want: false},
		{name: "both coordinates", latitude: 12.114, longitude: -86.236, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			address := Address{Latitude: tt.latitude, Longitude: tt.longitude}

			if got := address.HasCoordinates(); got != tt.want {
				t.Errorf("HasCoordinates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressIsComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		address Address
		want    bool
	}{
		{name: "all fields set", address: Address{Department: "Managua", Municipality: "Managua", AddressLine: "123 Main St"}, want: true},
		{name: "missing department", address: Address{Municipality: "Managua", AddressLine: "123 Main St"}, want: false},
		{name: "missing municipality", address: Address{Department: "Managua", AddressLine: "123 Main St"}, want: false},
		{name: "missing address line", address: Address{Department: "Managua", Municipality: "Managua"}, want: false},
		{name: "zero value", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.address.IsComplete(); got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}
