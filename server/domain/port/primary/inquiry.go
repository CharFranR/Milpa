package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
)

type InquiryUseCase interface {
	CreateInquiry(ctx context.Context, req dto.CreateInquiryRequest) (*dto.InquiryDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.InquiryDTO, error)
	GetByUser(ctx context.Context, UserId uuid.UUID) ([]*dto.InquiryDTO, error)
	UpdateInquiry(ctx context.Context, id uuid.UUID, req dto.UpdateInquiryRequest) error
}
