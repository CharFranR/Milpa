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

	review, err := domain.NewReview(principal.UserID, req.CompanyID, req.Rating, req.Comment, now)
	if err != nil {
		return nil, err
	}

	if err := uc.reviewRepo.Save(ctx, review); err != nil {
		return nil, err
	}

	return reviewToDTO(review), nil
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

var _ primary.ReviewUseCase = (*ReviewUseCaseImpl)(nil)

func reviewToDTO(review *domain.Review) *dto.ReviewDTO {
	return &dto.ReviewDTO{
		ID:        review.ID,
		UserID:    review.UserID,
		CompanyID: review.CompanyID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
	}
}
