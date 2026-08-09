package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		catName string
		wantErr error
	}{
		{name: "happy path", catName: "Agro"},
		{name: "empty name", wantErr: ErrNameRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			category, err := NewCategory(tt.catName)

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
			if category.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if category.Name != tt.catName {
				t.Errorf("name = %q, want %q", category.Name, tt.catName)
			}
			if category.Description != "" {
				t.Errorf("description = %q, want empty", category.Description)
			}
		})
	}
}
