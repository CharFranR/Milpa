package domain

import (
	"fmt"

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
