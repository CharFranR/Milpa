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

type ReviewUseCaseImpl struct {
	reviewRepo port.ReviewRepository
	timer      port.TimeProvider
}

func NewReviewUseCase(reviewRepo port.ReviewRepository, timer port.TimeProvider) *ReviewUseCaseImpl {
	return &ReviewUseCaseImpl{
		reviewRepo: reviewRepo,
		timer:      timer,
	}
}

func (uc *ReviewUseCaseImpl) CreateReview(ctx context.Context, req dto.CreateReviewRequest) (*dto.ReviewDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	now := uc.timer.Now()

	targetType, targetID, err := resolveReviewTarget(req)
	if err != nil {
		return nil, err
	}
	companyID := uuid.Nil
	if targetType == domain.ReviewTargetCompany {
		companyID = targetID
	}

	review, err := domain.NewReview(principal.UserID, targetType, targetID, companyID, req.Rating, req.Comment, now)
	if err != nil {
		return nil, err
	}

	if err := uc.reviewRepo.Save(ctx, review); err != nil {
		return nil, err
	}

	return reviewToDTO(review), nil
}

func resolveReviewTarget(req dto.CreateReviewRequest) (domain.ReviewTargetType, uuid.UUID, error) {
	hasLegacy := req.CompanyID != uuid.Nil
	hasExplicit := req.TargetID != uuid.Nil || req.TargetType != ""

	if hasLegacy && hasExplicit {
		return "", uuid.Nil, domain.ErrReviewTargetMismatch
	}
	if hasLegacy {
		return domain.ReviewTargetCompany, req.CompanyID, nil
	}
	if req.TargetID == uuid.Nil {
		return "", uuid.Nil, domain.ErrTargetRequired
	}

	targetType := domain.ReviewTargetType(req.TargetType)
	if !domain.ValidReviewTargetType(targetType) {
		return "", uuid.Nil, domain.ErrInvalidReviewTargetType
	}
	return targetType, req.TargetID, nil
}

func (uc *ReviewUseCaseImpl) FindByUser(ctx context.Context, userID uuid.UUID) ([]*dto.ReviewDTO, error) {
	reviews, err := uc.reviewRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ReviewDTO, len(reviews))
	for i := range reviews {
		dtos[i] = reviewToDTO(&reviews[i])
	}

	return dtos, nil
}

func (uc *ReviewUseCaseImpl) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]*dto.ReviewDTO, error) {
	reviews, err := uc.reviewRepo.FindByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ReviewDTO, len(reviews))
	for i := range reviews {
		dtos[i] = reviewToDTO(&reviews[i])
	}

	return dtos, nil
}

func (uc *ReviewUseCaseImpl) GetAverageRating(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) (*dto.ReviewAverageDTO, error) {
	if !domain.ValidReviewTargetType(targetType) {
		return nil, domain.ErrInvalidReviewTargetType
	}
	if targetID == uuid.Nil {
		return nil, domain.ErrTargetRequired
	}

	average, count, err := uc.reviewRepo.AverageRating(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}

	return &dto.ReviewAverageDTO{
		TargetType: string(targetType),
		TargetID:   targetID,
		Average:    average,
		Count:      count,
	}, nil
}

var _ primary.ReviewUseCase = (*ReviewUseCaseImpl)(nil)

func reviewToDTO(review *domain.Review) *dto.ReviewDTO {
	return &dto.ReviewDTO{
		ID:         review.ID,
		UserID:     review.AuthorID,
		CompanyID:  review.CompanyID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
		TargetType: string(review.TargetType),
		TargetID:   review.TargetID,
	}
}
