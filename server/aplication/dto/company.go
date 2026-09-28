package dto

import (
	"time"

	"github.com/google/uuid"
)

// CompanyView is what reading a company returns. Same boundary as UserView: the
// concrete type decides whether contact details travel with the response.
type CompanyView interface {
	companyView()
}

// PublicCompanyDTO is the marketplace representation of a company: its
// catalogue-facing identity, where it operates and whether it is verified. The
// address is reduced to the administrative area a buyer searches by; the street
// line, the phone and the mailbox are contact mechanism, not catalogue.
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

// PrivateCompanyDTO adds the contact card on top of the public representation.
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
