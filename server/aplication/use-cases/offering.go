package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type OfferingUseCaseImpl struct {
	offeringRepo  port.OfferingRepository
	fuzzyRetrival port.FuzzyRetrival
	userRepo      port.UserRepository
	timer         port.TimeProvider
}

func NewOfferingUseCase(offeringRepo port.OfferingRepository, userRepo port.UserRepository, timer port.TimeProvider, fuzzyRetrival port.FuzzyRetrival) *OfferingUseCaseImpl {
	return &OfferingUseCaseImpl{
		offeringRepo:  offeringRepo,
		userRepo:      userRepo,
		timer:         timer,
		fuzzyRetrival: fuzzyRetrival,
	}
}

func (uc *OfferingUseCaseImpl) CreateOffering(ctx context.Context, req dto.CreateOfferingRequest) (*dto.OfferingDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if user.ID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	now := uc.timer.Now()

	offering, err := domain.NewOffering(req.UserID, req.Name, req.Type, now)
	if err != nil {
		return nil, err
	}

	offering.Description = req.Description
	if req.Price > 0 {
		offering.Price = req.Price
	}
	offering.ImageURL = req.ImageURL

	if err := uc.offeringRepo.Save(ctx, offering); err != nil {
		return nil, err
	}

	// Save in elasticsearch
	err = uc.fuzzyRetrival.Index(ctx, &req)

	if err != nil {
		return nil, fmt.Errorf("CreateOffering.elasticsearch err: %w", err)
	}

	return offeringToDTO(offering), nil
}

func (uc *OfferingUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return offeringToDTO(offering), nil
}

func (uc *OfferingUseCaseImpl) GetByUserID(ctx context.Context, UserID uuid.UUID) ([]*dto.OfferingDTO, error) {
	offerings, err := uc.offeringRepo.FindByUserID(ctx, UserID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.OfferingDTO, len(offerings))
	for i := range offerings {
		dtos[i] = offeringToDTO(&offerings[i])
	}

	return dtos, nil
}

func (uc *OfferingUseCaseImpl) UpdateOffering(ctx context.Context, id uuid.UUID, req dto.UpdateOfferingRequest) error {
	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	now := uc.timer.Now()

	if req.Type != nil {
		offering.Type = *req.Type
	}
	if req.Name != nil {
		offering.Name = *req.Name
	}
	if req.Description != nil {
		offering.UpdateDescription(*req.Description, now)
	}
	if req.Price != nil {
		if err := offering.UpdatePrice(*req.Price, now); err != nil {
			return err
		}
	}
	if req.ImageURL != nil {
		offering.UpdateImage(*req.ImageURL, now)
	}

	offering.Touch(now)

	return uc.offeringRepo.Update(ctx, offering)
}

func (uc *OfferingUseCaseImpl) DeleteOffering(ctx context.Context, id uuid.UUID) error {

	err := uc.offeringRepo.Delete(ctx, id)

	if err != nil {
		return fmt.Errorf("Offering Delete error: %w", err)
	}

	err = uc.fuzzyRetrival.Delete(ctx, string(id.String()))

	if err != nil {
		return fmt.Errorf("Offering Fuzzy Delete error: %w", err)
	}

	return nil
}

var _ primary.OfferingUseCase = (*OfferingUseCaseImpl)(nil)

func offeringToDTO(offering *domain.Offering) *dto.OfferingDTO {
	return &dto.OfferingDTO{
		ID:          offering.ID,
		UserID:      offering.UserID,
		Type:        offering.Type,
		Name:        offering.Name,
		Description: offering.Description,
		Price:       offering.Price,
		ImageURL:    offering.ImageURL,
		CreatedAt:   offering.CreatedAt,
		UpdatedAt:   offering.UpdatedAt,
	}
}
