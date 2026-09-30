package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
)

const categoriesCachePrefix = "categories:"

type CachedCategoryUseCase struct {
	next  primary.CategoryUseCase
	cache port.Cache
}

func NewCachedCategoryUseCase(next primary.CategoryUseCase, cache port.Cache) *CachedCategoryUseCase {
	return &CachedCategoryUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedCategoryUseCase) GetAll(ctx context.Context) ([]*dto.CategoryDTO, error) {

	var categories []*dto.CategoryDTO

	_, err := uc.cache.Remember(
		ctx,
		categoriesCachePrefix+"all",
		5*time.Minute,
		&categories,
		func() error {
			result, err := uc.next.GetAll(ctx)
			if err != nil {
				return err
			}

			categories = result
			return nil
		},
	)

	return categories, err
}

func (uc *CachedCategoryUseCase) Create(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryDTO, error) {
	result, err := uc.next.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, categoriesCachePrefix)

	return result, nil
}

func (uc *CachedCategoryUseCase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCategoryRequest) (*dto.CategoryDTO, error) {
	result, err := uc.next.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, categoriesCachePrefix)

	return result, nil
}

func (uc *CachedCategoryUseCase) SetStatus(ctx context.Context, id uuid.UUID, req dto.CategoryStatusRequest) (*dto.CategoryDTO, error) {
	result, err := uc.next.SetStatus(ctx, id, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, categoriesCachePrefix)

	return result, nil
}

var _ primary.CategoryUseCase = (*CachedCategoryUseCase)(nil)
