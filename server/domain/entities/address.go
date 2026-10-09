package domain

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

type Address struct {
	ID           uuid.UUID
	Department   string
	Municipality string
	AddressLine  string
	Latitude     float64
	Longitude    float64
}

// Builder

func NewAddress(department, municipality, addressLine string) (*Address, error) {
	if department == "" {
		return nil, ErrDepartmentRequired
	}
	if municipality == "" {
		return nil, ErrMunicipalityRequired
	}
	if addressLine == "" {
		return nil, ErrAddressLineRequired
	}

	return &Address{
		ID:           uuid.New(),
		Department:   department,
		Municipality: municipality,
		AddressLine:  addressLine,
	}, nil
}

func (a Address) FullAddress() string {
	return fmt.Sprintf("%s, %s, %s", a.AddressLine, a.Municipality, a.Department)
}

func (a Address) HasCoordinates() bool {
	return a.Latitude != 0 && a.Longitude != 0
}

func (a Address) DistanceKM(other Address) (float64, bool) {
	if !a.HasCoordinates() || !other.HasCoordinates() {
		return 0, false
	}

	const earthRadiusKM = 6371.0

	lat1 := a.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - a.Latitude) * math.Pi / 180
	deltaLon := (other.Longitude - a.Longitude) * math.Pi / 180

	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h)), true
}

func (a Address) HasData() bool {
	return a.Department != "" || a.Municipality != "" || a.AddressLine != "" || a.HasCoordinates()
}

func (a Address) ValidateCoordinates() error {
	if a.Latitude < -90 || a.Latitude > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidInput)
	}
	if a.Longitude < -180 || a.Longitude > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidInput)
	}
	return nil
}

func (a Address) IsComplete() bool {
	return a.Department != "" && a.Municipality != "" && a.AddressLine != ""
}
