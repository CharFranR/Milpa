package usecases

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type CompanyUseCaseImpl struct {
	companyRepo  port.CompanyRepository
	userRepo     port.UserRepository
	categoryRepo port.CategoryRepository
	timer        port.TimeProvider
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

func (uc *CompanyUseCaseImpl) CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.PrivateCompanyDTO, error) {
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

	// The caller owns the company it just created, so this is the private view.
	return privateCompanyDTO(company), nil
}

// GetByID returns the representation the caller is entitled to: the contact card
// for the owner and for an admin, the public listing for everyone else.
func (uc *CompanyUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.CompanyView, error) {
	company, err := uc.companyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return companyViewFor(ctx, company), nil
}

// GetByOwner lists the companies an owner registered, under the same boundary
// as GetByID. This read is unauthenticated too, so leaving it returning contact
// details would hand out exactly what GetByID refuses to.
func (uc *CompanyUseCaseImpl) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]dto.CompanyView, error) {
	companies, err := uc.companyRepo.FindByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.CompanyView, len(companies))
	for i := range companies {
		dtos[i] = companyViewFor(ctx, &companies[i])
	}

	return dtos, nil
}

func (uc *CompanyUseCaseImpl) UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	company, err := uc.companyRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if company.Owner.ID != principal.UserID {
		return domain.ErrForbidden
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

// companyViewFor applies the contact-detail boundary. See userViewFor: the
// owner and an admin get the contact card, everyone else gets the listing.
func companyViewFor(ctx context.Context, company *domain.Company) dto.CompanyView {
	principal, ok := auth.FromContext(ctx)
	if ok && (principal.UserID == company.Owner.ID || principal.Role == domain.RoleAdmin) {
		return privateCompanyDTO(company)
	}
	return publicCompanyDTO(company)
}

func publicCompanyDTO(company *domain.Company) *dto.PublicCompanyDTO {
	var categoryID uuid.UUID
	if len(company.Category) > 0 {
		categoryID = company.Category[0].ID
	}

	return &dto.PublicCompanyDTO{
		ID:           company.ID,
		Name:         company.Name,
		CategoryID:   categoryID,
		OwnerID:      company.Owner.ID,
		Department:   company.Address.Department,
		Municipality: company.Address.Municipality,
		Description:  company.Description,
		Website:      company.Website,
		Verified:     company.Verified,
		CreatedAt:    company.CreatedAt,
		UpdatedAt:    company.UpdatedAt,
	}
}

func privateCompanyDTO(company *domain.Company) *dto.PrivateCompanyDTO {
	return &dto.PrivateCompanyDTO{
		PublicCompanyDTO: *publicCompanyDTO(company),
		Email:            company.Email,
		PhoneNumber:      company.PhoneNumber,
		AddressLine:      company.Address.AddressLine,
	}
}
