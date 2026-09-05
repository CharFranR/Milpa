package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type CompanyUseCase interface {
	CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.CompanyDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CompanyDTO, error)
	GetByOwner(ctx context.Context, OwnerId uuid.UUID) ([]*dto.CompanyDTO, error)
	UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error
}
