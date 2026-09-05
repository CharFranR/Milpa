package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOffering(t *testing.T) {
	t.Parallel()

	now := time.Now()
	companyID := uuid.New()

	tests := []struct {
		name         string
		companyID    uuid.UUID
		offeringName string
		offeringType OfferingType
		now          time.Time
		wantErr      error
	}{
		{name: "happy path product", companyID: companyID, offeringName: "Corn", offeringType: OfferingProduct, now: now},
		{name: "happy path service", companyID: companyID, offeringName: "Transport", offeringType: OfferingService, now: now},
		{name: "empty name", companyID: companyID, offeringType: OfferingProduct, now: now, wantErr: ErrNameRequired},
		{name: "invalid type", companyID: companyID, offeringName: "Corn", offeringType: OfferingType(99), now: now, wantErr: ErrInvalidOfferingType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offering, err := NewOffering(tt.companyID, tt.offeringName, tt.offeringType, tt.now)

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
			if offering.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if offering.CompanyID != tt.companyID {
				t.Errorf("company id = %v, want %v", offering.CompanyID, tt.companyID)
			}
			if offering.Type != tt.offeringType {
				t.Errorf("type = %v, want %v", offering.Type, tt.offeringType)
			}
			if offering.Name != tt.offeringName {
				t.Errorf("name = %q, want %q", offering.Name, tt.offeringName)
			}
			if offering.Description != "" {
				t.Errorf("description = %q, want empty", offering.Description)
			}
			if offering.Price != 0 {
				t.Errorf("price = %v, want zero value", offering.Price)
			}
			if offering.ImageURL != "" {
				t.Errorf("image url = %q, want empty", offering.ImageURL)
			}
			if !offering.CreatedAt.Equal(tt.now) {
				t.Errorf("created at = %v, want %v", offering.CreatedAt, tt.now)
			}
			if !offering.UpdatedAt.Equal(tt.now) {
				t.Errorf("updated at = %v, want %v", offering.UpdatedAt, tt.now)
			}
		})
	}
}

func TestOfferingIsProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  OfferingType
		want bool
	}{
		{name: "product type", typ: OfferingProduct, want: true},
		{name: "service type", typ: OfferingService, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offering := Offering{Type: tt.typ}

			if got := offering.IsProduct(); got != tt.want {
				t.Errorf("IsProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOfferingIsService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  OfferingType
		want bool
	}{
		{name: "service type", typ: OfferingService, want: true},
		{name: "product type", typ: OfferingProduct, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offering := Offering{Type: tt.typ}

			if got := offering.IsService(); got != tt.want {
				t.Errorf("IsService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOfferingUpdatePrice(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name      string
		price     float64
		updatedAt time.Time
		wantPrice float64
		wantErr   error
	}{
		{name: "valid price", price: 25.5, updatedAt: earlier, wantPrice: 25.5},
		{name: "zero price", price: 0, updatedAt: earlier, wantPrice: 0, wantErr: ErrInvalidPrice},
		{name: "negative price", price: -1, updatedAt: earlier, wantPrice: 0, wantErr: ErrInvalidPrice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offering := &Offering{Price: 0, UpdatedAt: tt.updatedAt}

			err := offering.UpdatePrice(tt.price, now)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if offering.Price != tt.wantPrice {
					t.Errorf("price = %v, want %v", offering.Price, tt.wantPrice)
				}
				if !offering.UpdatedAt.Equal(tt.updatedAt) {
					t.Errorf("updated at = %v, want %v", offering.UpdatedAt, tt.updatedAt)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if offering.Price != tt.wantPrice {
				t.Errorf("price = %v, want %v", offering.Price, tt.wantPrice)
			}
			if !offering.UpdatedAt.Equal(now) {
				t.Errorf("updated at = %v, want %v", offering.UpdatedAt, now)
			}
		})
	}
}

func TestOfferingUpdateDescription(t *testing.T) {
	t.Parallel()

	now := time.Now()
	offering := &Offering{}

	offering.UpdateDescription("Fresh corn", now)

	if offering.Description != "Fresh corn" {
		t.Errorf("Description = %q, want %q", offering.Description, "Fresh corn")
	}
	if !offering.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", offering.UpdatedAt, now)
	}
}

func TestOfferingUpdateImage(t *testing.T) {
	t.Parallel()

	now := time.Now()
	offering := &Offering{}

	offering.UpdateImage("https://example.com/corn.jpg", now)

	if offering.ImageURL != "https://example.com/corn.jpg" {
		t.Errorf("ImageURL = %q, want %q", offering.ImageURL, "https://example.com/corn.jpg")
	}
	if !offering.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", offering.UpdatedAt, now)
	}
}

func TestOfferingTouch(t *testing.T) {
	t.Parallel()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name     string
		offering Offering
	}{
		{name: "zero value", offering: Offering{}},
		{name: "updates previous timestamp", offering: Offering{UpdatedAt: earlier}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offering := tt.offering
			offering.Touch(now)

			if !offering.UpdatedAt.Equal(now) {
				t.Errorf("UpdatedAt = %v, want %v", offering.UpdatedAt, now)
			}
		})
	}
}

func TestOfferingTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  OfferingType
		want string
	}{
		{name: "product", typ: OfferingProduct, want: "product"},
		{name: "service", typ: OfferingService, want: "service"},
		{name: "unknown value", typ: OfferingType(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.typ.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
