package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

type CachedInquiryUseCase struct {
	next  primary.InquiryUseCase
	cache port.Cache
}

func NewCachedInquiryUseCase(next primary.InquiryUseCase, cache port.Cache) *CachedInquiryUseCase {
	return &CachedInquiryUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedInquiryUseCase) CreateInquiry(ctx context.Context, req dto.CreateInquiryRequest) (*dto.InquiryDTO, error) {
	result, err := uc.next.CreateInquiry(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.Delete(ctx, "inquiries:byuser:"+result.UserID.String())

	return result, nil
}

func (uc *CachedInquiryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.InquiryDTO, error) {
	var inquiry *dto.InquiryDTO

	_, err := uc.cache.Remember(
		ctx,
		"inquiry:"+id.String(),
		5*time.Minute,
		&inquiry,
		func() error {
			result, err := uc.next.GetByID(ctx, id)
			if err != nil {
				return err
			}

			inquiry = result
			return nil
		},
	)

	return inquiry, err
}

func (uc *CachedInquiryUseCase) GetByUser(ctx context.Context, userID uuid.UUID) ([]*dto.InquiryDTO, error) {
	var inquiries []*dto.InquiryDTO

	_, err := uc.cache.Remember(
		ctx,
		"inquiries:byuser:"+userID.String(),
		5*time.Minute,
		&inquiries,
		func() error {
			result, err := uc.next.GetByUser(ctx, userID)
			if err != nil {
				return err
			}

			inquiries = result
			return nil
		},
	)

	return inquiries, err
}

func (uc *CachedInquiryUseCase) UpdateInquiry(ctx context.Context, id uuid.UUID, req dto.UpdateInquiryRequest) error {
	err := uc.next.UpdateInquiry(ctx, id, req)
	if err != nil {
		return err
	}

	_ = uc.cache.Delete(ctx, "inquiry:"+id.String())

	return nil
}

var _ primary.InquiryUseCase = (*CachedInquiryUseCase)(nil)
