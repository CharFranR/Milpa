package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type CompanyUseCase interface {
	CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.PrivateCompanyDTO, error)
	// GetByID and GetByOwner return dto.PublicCompanyDTO or
	// dto.PrivateCompanyDTO depending on whether the caller owns the company or
	// is an admin.
	GetByID(ctx context.Context, id uuid.UUID) (dto.CompanyView, error)
	GetByOwner(ctx context.Context, OwnerId uuid.UUID) ([]dto.CompanyView, error)
	UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error
}
