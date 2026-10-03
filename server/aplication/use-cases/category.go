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

type CategoryUseCaseImpl struct {
	categoryRepo port.CategoryRepository
}

func NewCategoryUseCase(categoryRepo port.CategoryRepository) *CategoryUseCaseImpl {
	return &CategoryUseCaseImpl{
		categoryRepo: categoryRepo,
	}
}

func (uc *CategoryUseCaseImpl) GetAll(ctx context.Context) ([]*dto.CategoryDTO, error) {
	categories, err := uc.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.CategoryDTO, 0, len(categories))
	for i := range categories {
		if !categories[i].IsActive {
			continue
		}
		dtos = append(dtos, categoryToDTO(&categories[i]))
	}

	return dtos, nil
}

func (uc *CategoryUseCaseImpl) Create(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	category, err := domain.NewCategory(req.Name)
	if err != nil {
		return nil, err
	}

	category.Description = req.Description
	category.MainCategory = resolveMainCategory(req.MainCategory)
	category.DefaultUnitOfMeasureID = req.DefaultUnitOfMeasureID

	if err := uc.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}

	return categoryToDTO(category), nil
}

func (uc *CategoryUseCaseImpl) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCategoryRequest) (*dto.CategoryDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	category, err := uc.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, domain.ErrNameRequired
		}
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.MainCategory != nil {
		category.MainCategory = resolveMainCategory(*req.MainCategory)
	}
	if req.DefaultUnitOfMeasureID != nil {
		category.DefaultUnitOfMeasureID = req.DefaultUnitOfMeasureID
	}

	if err := uc.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}

	return categoryToDTO(category), nil
}

func (uc *CategoryUseCaseImpl) SetStatus(ctx context.Context, id uuid.UUID, req dto.CategoryStatusRequest) (*dto.CategoryDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	category, err := uc.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.IsActive {
		category.Activate()
	} else {
		category.Deactivate()
	}

	if err := uc.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}

	return categoryToDTO(category), nil
}

var _ primary.CategoryUseCase = (*CategoryUseCaseImpl)(nil)

func resolveMainCategory(mainCategory string) string {
	if mainCategory == "" {
		return domain.DefaultMainCategory
	}
	return mainCategory
}

func categoryToDTO(category *domain.Category) *dto.CategoryDTO {
	return &dto.CategoryDTO{
		ID:                     category.ID,
		Name:                   category.Name,
		Description:            category.Description,
		MainCategory:           category.MainCategory,
		IsActive:               category.IsActive,
		DefaultUnitOfMeasureID: category.DefaultUnitOfMeasureID,
	}
}
