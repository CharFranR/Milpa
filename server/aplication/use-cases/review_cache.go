package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"

	"github.com/google/uuid"
)

// averageKey is the aggregate's own cache key, keyed by target rather than by
// company: a review of a user moves that user's average and no company's, so a
// key that could not name a user would be served a company's average for a
// farmer.
func averageKey(targetType domain.ReviewTargetType, targetID uuid.UUID) string {
	return "reviews:avg:" + string(targetType) + ":" + targetID.String()
}

type CachedReviewUseCase struct {
	next  primary.ReviewUseCase
	cache port.Cache
}

func NewCachedReviewUseCase(next primary.ReviewUseCase, cache port.Cache) *CachedReviewUseCase {
	return &CachedReviewUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedReviewUseCase) CreateReview(ctx context.Context, req dto.CreateReviewRequest) (*dto.ReviewDTO, error) {
	result, err := uc.next.CreateReview(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.Delete(ctx, "reviews:byuser:"+result.UserID.String())
	// Only a company review appears in a company's list, and only a company
	// review has a company to mirror.
	if result.TargetType == string(domain.ReviewTargetCompany) {
		_ = uc.cache.Delete(ctx, "reviews:bycompany:"+result.CompanyID.String())
	}
	_ = uc.cache.Delete(ctx, averageKey(domain.ReviewTargetType(result.TargetType), result.TargetID))
	_ = uc.cache.Delete(ctx, "reviews:bytarget:"+result.TargetType+":"+result.TargetID.String())

	return result, nil
}

func (uc *CachedReviewUseCase) FindByUser(ctx context.Context, userID uuid.UUID) ([]*dto.ReviewDTO, error) {
	var reviews []*dto.ReviewDTO

	_, err := uc.cache.Remember(
		ctx,
		"reviews:byuser:"+userID.String(),
		5*time.Minute,
		&reviews,
		func() error {
			result, err := uc.next.FindByUser(ctx, userID)
			if err != nil {
				return err
			}

			reviews = result
			return nil
		},
	)

	return reviews, err
}

func (uc *CachedReviewUseCase) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]*dto.ReviewDTO, error) {
	var reviews []*dto.ReviewDTO

	_, err := uc.cache.Remember(
		ctx,
		"reviews:bycompany:"+companyID.String(),
		5*time.Minute,
		&reviews,
		func() error {
			result, err := uc.next.FindByCompany(ctx, companyID)
			if err != nil {
				return err
			}

			reviews = result
			return nil
		},
	)

	return reviews, err
}

func (uc *CachedReviewUseCase) FindByTarget(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) ([]*dto.ReviewDTO, error) {
	var reviews []*dto.ReviewDTO

	_, err := uc.cache.Remember(
		ctx,
		"reviews:bytarget:"+string(targetType)+":"+targetID.String(),
		5*time.Minute,
		&reviews,
		func() error {
			result, err := uc.next.FindByTarget(ctx, targetType, targetID)
			if err != nil {
				return err
			}

			reviews = result
			return nil
		},
	)

	return reviews, err
}

// The aggregate is read through the decorator rather than around it, on the
// same five-minute TTL as the lists beside it.
func (uc *CachedReviewUseCase) GetAverageRating(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) (*dto.ReviewAverageDTO, error) {
	var average dto.ReviewAverageDTO

	_, err := uc.cache.Remember(
		ctx,
		averageKey(targetType, targetID),
		5*time.Minute,
		&average,
		func() error {
			result, err := uc.next.GetAverageRating(ctx, targetType, targetID)
			if err != nil {
				return err
			}

			average = *result
			return nil
		},
	)

	if err != nil {
		return nil, err
	}
	return &average, nil
}

var _ primary.ReviewUseCase = (*CachedReviewUseCase)(nil)
