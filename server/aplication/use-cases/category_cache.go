package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
)

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
		"categories:all",
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

var _ primary.CategoryUseCase = (*CachedCategoryUseCase)(nil)
