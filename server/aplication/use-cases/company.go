package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/CharFranR/Hackaton2026/aplication/dto"
	domain "github.com/CharFranR/Hackaton2026/domain/entities"
	"github.com/CharFranR/Hackaton2026/domain/port/primary"
	"github.com/CharFranR/Hackaton2026/domain/port/secondary"
	"github.com/CharFranR/Hackaton2026/internal/auth"
)

type CompanyUseCaseImpl struct {
	companyRepo port.CompanyRepository
	userRepo    port.UserRepository
	categoryRepo port.CategoryRepository
	timer       port.TimeProvider
}

func NewCompanyUseCase(
	companyRepo port.CompanyRepository,
	userRepo port.UserRepository,
	categoryRepo port.CategoryRepository,
	timer port.TimeProvider,
) *CompanyUseCaseImpl {
	return &CompanyUseCaseImpl{
		companyRepo:  companyRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		timer:        timer,
	}
}

func (uc *CompanyUseCaseImpl) CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.CompanyDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	now := uc.timer.Now()

	company, err := domain.NewCompany(domain.User{ID: principal.UserID}, req.Name, now)
	if err != nil {
		return nil, err
	}

	if req.CategoryID != uuid.Nil {
		category, err := uc.categoryRepo.FindByID(ctx, req.CategoryID)
		if err != nil {
			return nil, err
		}
		company.Category = []domain.Category{*category}
	}

	company.Address = domain.Address{AddressLine: req.Address}
	company.Description = req.Description
	company.PhoneNumber = req.PhoneNumber
	company.Email = req.Email
	company.Website = req.Website

	if err := uc.companyRepo.Save(ctx, company); err != nil {
		return nil, err
	}

	return companyToDTO(company), nil
}

func (uc *CompanyUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.CompanyDTO, error) {
	company, err := uc.companyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return companyToDTO(company), nil
}

func (uc *CompanyUseCaseImpl) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*dto.CompanyDTO, error) {
	companies, err := uc.companyRepo.FindByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.CompanyDTO, len(companies))
	for i := range companies {
		dtos[i] = companyToDTO(&companies[i])
	}

	return dtos, nil
}

func (uc *CompanyUseCaseImpl) UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error {
	company, err := uc.companyRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != nil {
		company.Name = *req.Name
	}
	if req.Address != nil {
		company.Address = domain.Address{AddressLine: *req.Address}
	}
	if req.Description != nil {
		company.Description = *req.Description
	}
	if req.PhoneNumber != nil {
		company.PhoneNumber = *req.PhoneNumber
	}
	if req.Email != nil {
		company.Email = *req.Email
	}
	if req.Website != nil {
		company.Website = *req.Website
	}

	company.Touch(uc.timer.Now())

	return uc.companyRepo.Update(ctx, company)
}

var _ primary.CompanyUseCase = (*CompanyUseCaseImpl)(nil)

func companyToDTO(company *domain.Company) *dto.CompanyDTO {
	var categoryID uuid.UUID
	if len(company.Category) > 0 {
		categoryID = company.Category[0].ID
	}

	return &dto.CompanyDTO{
		ID:          company.ID,
		Name:        company.Name,
		CategoryID:  categoryID,
		OwnerID:     company.Owner.ID,
		Address:     company.Address.FullAddress(),
		Description: company.Description,
		PhoneNumber: company.PhoneNumber,
		Email:       company.Email,
		Website:     company.Website,
		Verified:    company.Verified,
		CreatedAt:   company.CreatedAt,
		UpdatedAt:   company.UpdatedAt,
	}
}
