package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/CharFranR/Hackaton2026/aplication/dto"
	domain "github.com/CharFranR/Hackaton2026/domain/entities"
	"github.com/CharFranR/Hackaton2026/domain/port/primary"
	port "github.com/CharFranR/Hackaton2026/domain/port/secondary"
	"github.com/CharFranR/Hackaton2026/internal/auth"
)

type InquiryUseCaseImpl struct {
	inquiryRepo port.InquiryRepository
	timer       port.TimeProvider
}

func NewInquiryUseCase(inquiryRepo port.InquiryRepository, timer port.TimeProvider) *InquiryUseCaseImpl {
	return &InquiryUseCaseImpl{
		inquiryRepo: inquiryRepo,
		timer:       timer,
	}
}

func (uc *InquiryUseCaseImpl) CreateInquiry(ctx context.Context, req dto.CreateInquiryRequest) (*dto.InquiryDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	now := uc.timer.Now()

	inquiry, err := domain.NewInquiry(principal.UserID, req.OfferingID, req.Message, now)
	if err != nil {
		return nil, err
	}

	if err := uc.inquiryRepo.Save(ctx, inquiry); err != nil {
		return nil, err
	}

	return inquiryToDTO(inquiry), nil
}

func (uc *InquiryUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.InquiryDTO, error) {
	inquiry, err := uc.inquiryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return inquiryToDTO(inquiry), nil
}

func (uc *InquiryUseCaseImpl) GetByUser(ctx context.Context, userID uuid.UUID) ([]*dto.InquiryDTO, error) {
	inquiries, err := uc.inquiryRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.InquiryDTO, len(inquiries))
	for i := range inquiries {
		dtos[i] = inquiryToDTO(&inquiries[i])
	}

	return dtos, nil
}

func (uc *InquiryUseCaseImpl) UpdateInquiry(ctx context.Context, id uuid.UUID, req dto.UpdateInquiryRequest) error {
	inquiry, err := uc.inquiryRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Status != nil {
		switch *req.Status {
		case domain.InquiryRead:
			inquiry.MarkAsRead()
		case domain.InquiryReplied:
			inquiry.MarkAsReplied()
		case domain.InquiryClosed:
			inquiry.Close()
		default:
			return domain.ErrInvalidInput
		}
	}

	return uc.inquiryRepo.Update(ctx, inquiry)
}

var _ primary.InquiryUseCase = (*InquiryUseCaseImpl)(nil)

func inquiryToDTO(inquiry *domain.Inquiry) *dto.InquiryDTO {
	return &dto.InquiryDTO{
		ID:         inquiry.ID,
		UserID:     inquiry.UserID,
		OfferingID: inquiry.OfferingID,
		Message:    inquiry.Message,
		Status:     inquiry.Status,
		CreatedAt:  inquiry.CreatedAt,
		UpdatedAt:  inquiry.CreatedAt,
	}
}
