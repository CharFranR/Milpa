package primary

import (
	"context"

	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type ReviewUseCase interface {
	CreateReview(ctx context.Context, req dto.CreateReviewRequest) (*dto.ReviewDTO, error)
	FindByUser(ctx context.Context, UserId uuid.UUID) ([]*dto.ReviewDTO, error)
	FindByCompany(ctx context.Context, CompanyId uuid.UUID) ([]*dto.ReviewDTO, error)
}
