package entities_test

import (
	"math"
	"testing"

	domain "milpa/domain/entities"
)

func TestAddressDistanceKM(t *testing.T) {
	t.Parallel()

	leon := domain.Address{Latitude: 12.4379, Longitude: -86.8781}
	managua := domain.Address{Latitude: 12.1149926, Longitude: -86.2361744}

	t.Run("known pair is within tolerance", func(t *testing.T) {
		t.Parallel()

		got, ok := leon.DistanceKM(managua)
		if !ok {
			t.Fatal("DistanceKM = unknown for two addresses with coordinates")
		}
		if math.Abs(got-78.45) > 0.5 {
			t.Errorf("distance = %v, want 78.45 +/- 0.5", got)
		}
	})

	t.Run("same point is zero", func(t *testing.T) {
		t.Parallel()

		got, ok := managua.DistanceKM(managua)
		if !ok {
			t.Fatal("DistanceKM = unknown for an address with coordinates")
		}
		if got != 0 {
			t.Errorf("distance = %v, want 0", got)
		}
	})

	t.Run("missing coordinates are unknown", func(t *testing.T) {
		t.Parallel()

		missing := domain.Address{Department: "Leon", Municipality: "Leon"}
		zeroLatitude := domain.Address{Latitude: 0, Longitude: -86.8781}

		if _, ok := leon.DistanceKM(missing); ok {
			t.Error("DistanceKM = known when the other address has no coordinates")
		}
		if _, ok := missing.DistanceKM(leon); ok {
			t.Error("DistanceKM = known when the receiver has no coordinates")
		}
		if _, ok := leon.DistanceKM(zeroLatitude); ok {
			t.Error("DistanceKM = known for a zero latitude")
		}
	})
}
