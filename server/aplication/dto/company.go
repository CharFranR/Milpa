package dto

import (
	"time"

	"github.com/google/uuid"
)

type CompanyView interface {
	companyView()
}

type PublicCompanyDTO struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CategoryID   uuid.UUID `json:"category_id"`
	OwnerID      uuid.UUID `json:"owner_id"`
	Department   string    `json:"department"`
	Municipality string    `json:"municipality"`
	Description  string    `json:"description"`
	Website      string    `json:"website"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PrivateCompanyDTO struct {
	PublicCompanyDTO
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	AddressLine string `json:"address_line"`
}

func (PublicCompanyDTO) companyView()  {}
func (PrivateCompanyDTO) companyView() {}

type RegisterCompanyRequest struct {
	Name        string    `json:"name"`
	CategoryID  uuid.UUID `json:"category_id,omitempty"`
	Address     string    `json:"address,omitempty"`
	Description string    `json:"description,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Email       string    `json:"email,omitempty"`
	Website     string    `json:"website,omitempty"`
}

type UpdateCompanyRequest struct {
	Name        *string `json:"name,omitempty"`
	Address     *string `json:"address,omitempty"`
	Description *string `json:"description,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Email       *string `json:"email,omitempty"`
	Website     *string `json:"website,omitempty"`
}
