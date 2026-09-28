package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

type CachedCompanyUseCase struct {
	next  primary.CompanyUseCase
	cache port.Cache
}

func NewCachedCompanyUseCase(next primary.CompanyUseCase, cache port.Cache) *CachedCompanyUseCase {
	return &CachedCompanyUseCase{
		next:  next,
		cache: cache,
	}
}

// cachedCompanyView is the cache envelope for a company read; see
// cachedUserView for why the union is stored rather than one of the views.
type cachedCompanyView struct {
	Public  *dto.PublicCompanyDTO  `json:"public"`
	Private *dto.PrivateCompanyDTO `json:"private"`
}

func cacheCompanyView(view dto.CompanyView) cachedCompanyView {
	switch v := view.(type) {
	case *dto.PrivateCompanyDTO:
		return cachedCompanyView{Private: v}
	default:
		public, _ := view.(*dto.PublicCompanyDTO)
		return cachedCompanyView{Public: public}
	}
}

func (uc *CachedCompanyUseCase) GetByID(ctx context.Context, id uuid.UUID) (dto.CompanyView, error) {
	var cached cachedCompanyView

	_, err := uc.cache.Remember(
		ctx,
		viewerCacheKey(ctx, "company:", id),
		5*time.Minute,
		&cached,
		func() error {
			result, err := uc.next.GetByID(ctx, id)
			if err != nil {
				return err
			}

			cached = cacheCompanyView(result)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if cached.Private != nil {
		return cached.Private, nil
	}
	if cached.Public != nil {
		return cached.Public, nil
	}
	return nil, nil
}

func (uc *CachedCompanyUseCase) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]dto.CompanyView, error) {
	var cached []cachedCompanyView

	key := viewerCacheKey(ctx, "company:byowner:", ownerID)

	_, err := uc.cache.Remember(
		ctx,
		key,
		5*time.Minute,
		&cached,
		func() error {
			results, err := uc.next.GetByOwner(ctx, ownerID)
			if err != nil {
				return err
			}

			cached = make([]cachedCompanyView, len(results))
			for i := range results {
				cached[i] = cacheCompanyView(results[i])
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	views := make([]dto.CompanyView, 0, len(cached))
	for i := range cached {
		switch {
		case cached[i].Private != nil:
			views = append(views, cached[i].Private)
		case cached[i].Public != nil:
			views = append(views, cached[i].Public)
		}
	}

	return views, nil
}

func (uc *CachedCompanyUseCase) CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.PrivateCompanyDTO, error) {
	result, err := uc.next.CreateCompany(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "company:byowner:"+result.OwnerID.String()+":")

	return result, nil
}

func (uc *CachedCompanyUseCase) UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error {
	err := uc.next.UpdateCompany(ctx, id, req)
	if err != nil {
		return err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "company:"+id.String()+":")

	return err
}

var _ primary.CompanyUseCase = (*CachedCompanyUseCase)(nil)
