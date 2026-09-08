package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

type CachedOfferingUseCase struct {
	next  primary.OfferingUseCase
	cache port.Cache
}

func NewCachedOfferingUseCase(next primary.OfferingUseCase, cache port.Cache) *CachedOfferingUseCase {
	return &CachedOfferingUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedOfferingUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	var offering *dto.OfferingDTO

	_, err := uc.cache.Remember(
		ctx,
		"offering:"+id.String(),
		5*time.Minute,
		&offering,
		func() error {
			result, err := uc.next.GetByID(ctx, id)
			if err != nil {
				return err
			}
			offering = result
			return nil
		},
	)

	return offering, err
}

func (uc *CachedOfferingUseCase) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]*dto.OfferingDTO, error) {
	var offering []*dto.OfferingDTO

	_, err := uc.cache.Remember(
		ctx,
		"offerings:bycompany:"+companyID.String(),
		5*time.Minute,
		&offering,
		func() error {
			result, err := uc.next.GetByCompany(ctx, companyID)
			if err != nil {
				return err
			}
			offering = result
			return nil
		},
	)

	return offering, err
}

func (uc *CachedOfferingUseCase) CreateOffering(ctx context.Context, req dto.CreateOfferingRequest) (*dto.OfferingDTO, error) {
	result, err := uc.next.CreateOffering(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.Delete(ctx, "offerings:bycompany:"+result.CompanyID.String())

	return result, nil
}

func (uc *CachedOfferingUseCase) UpdateOffering(ctx context.Context, id uuid.UUID, req dto.UpdateOfferingRequest) error {
	err := uc.next.UpdateOffering(ctx, id, req)
	if err != nil {
		return err
	}

	_ = uc.cache.Delete(ctx, "offering:"+id.String())

	return nil
}

var _ primary.OfferingUseCase = (*CachedOfferingUseCase)(nil)
