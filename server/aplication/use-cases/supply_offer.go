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

type SupplyOfferUseCaseImpl struct {
	supplyOfferRepo   port.SupplyOfferRepository
	supplyRequestRepo port.SupplyRequestRepository
	matchRepo         port.MatchRepository
	timer             port.TimeProvider
}

func NewSupplyOfferUseCase(supplyOfferRepo port.SupplyOfferRepository, supplyRequestRepo port.SupplyRequestRepository, matchRepo port.MatchRepository, timer port.TimeProvider) *SupplyOfferUseCaseImpl {
	return &SupplyOfferUseCaseImpl{
		supplyOfferRepo:   supplyOfferRepo,
		supplyRequestRepo: supplyRequestRepo,
		matchRepo:         matchRepo,
		timer:             timer,
	}
}

func (uc *SupplyOfferUseCaseImpl) Create(ctx context.Context, req dto.SupplyOfferDTO) (*dto.SupplyOfferDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if !isFarmer(principal) {
		return nil, domain.ErrForbidden
	}

	if req.SupplyRequest == nil || *req.SupplyRequest == uuid.Nil {
		return nil, fmt.Errorf("%w: supply request id is required", domain.ErrInvalidInput)
	}
	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("%w: total amount must be greater than zero", domain.ErrInvalidInput)
	}

	if req.PricePerUnit == nil || *req.PricePerUnit <= 0 {
		return nil, domain.ErrInvalidPrice
	}
	supplyRequestID := *req.SupplyRequest

	supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}
	if !supplyRequest.IsOpen() {
		return nil, domain.ErrInvalidRequestStatus
	}
	if supplyRequest.BuyerID == principal.UserID {
		return nil, fmt.Errorf("%w: cannot create an offer on your own supply request", domain.ErrForbidden)
	}

	if _, err := uc.supplyOfferRepo.FindBySupplierAndRequest(ctx, principal.UserID, supplyRequestID); err == nil {
		return nil, domain.ErrDuplicate
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	if !supplyRequest.MultipleProviders {
		hasActiveMatch, err := uc.matchRepo.ExistsActiveByRequest(ctx, supplyRequestID)
		if err != nil {
			return nil, err
		}
		if hasActiveMatch {
			return nil, primary.ErrActiveMatch
		}
	}

	if err := validateOfferAgainstRequest(supplyRequest, req.TotalAmount); err != nil {
		return nil, err
	}

	supplyOffer := domain.NewSupplyOffer(
		principal.UserID, supplyRequestID, req.TotalAmount, req.AmountUnit,
		req.ProposedDeliveryDay, req.DeliveryAvailable,
	)

	supplyOffer.PricePerUnit = req.PricePerUnit
	supplyOffer.Comments = req.Comments

	if err := uc.supplyOfferRepo.Create(ctx, supplyOffer); err != nil {
		return nil, err
	}

	return supplyOfferToDTO(supplyOffer), nil
}

func (uc *SupplyOfferUseCaseImpl) Update(ctx context.Context, id uuid.UUID, req dto.SupplyOfferUpdateDTO) error {
	supplyOffer, err := uc.getOwnedActionableOffer(ctx, id)
	if err != nil {
		return err
	}
	if req.TotalAmount <= 0 {
		return fmt.Errorf("%w: total amount must be greater than zero", domain.ErrInvalidInput)
	}
	if req.PricePerUnit == nil || *req.PricePerUnit <= 0 {
		return domain.ErrInvalidPrice
	}

	supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, supplyOffer.SupplyRequest)
	if err != nil {
		return err
	}
	if err := validateOfferAgainstRequest(supplyRequest, req.TotalAmount); err != nil {
		return err
	}

	supplyOffer.TotalAmount = req.TotalAmount
	supplyOffer.AmountUnit = req.AmountUnit
	supplyOffer.PricePerUnit = req.PricePerUnit
	supplyOffer.Comments = req.Comments
	supplyOffer.ProposedDeliveryDay = req.ProposedDeliveryDay
	supplyOffer.DeliveryAvailable = req.DeliveryAvailable
	supplyOffer.UpdatedAt = uc.timer.Now()

	return uc.supplyOfferRepo.Update(ctx, &supplyOffer)
}

func (uc *SupplyOfferUseCaseImpl) Withdraw(ctx context.Context, id uuid.UUID) error {
	supplyOffer, err := uc.getOwnedActionableOffer(ctx, id)
	if err != nil {
		return err
	}

	if err := supplyOffer.Withdraw(); err != nil {
		return err
	}

	return uc.supplyOfferRepo.Update(ctx, &supplyOffer)
}

func (uc *SupplyOfferUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyOfferDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: null supply offer id", domain.ErrInvalidInput)
	}

	supplyOffer, err := uc.supplyOfferRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if supplyOffer.SupplierID != principal.UserID {
		supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, supplyOffer.SupplyRequest)
		if err != nil {
			return nil, err
		}
		if supplyRequest.BuyerID != principal.UserID {
			return nil, domain.ErrForbidden
		}
	}

	return supplyOfferToDTO(&supplyOffer), nil
}

func (uc *SupplyOfferUseCaseImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.SupplyOfferDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if supplyRequestID == uuid.Nil {
		return nil, fmt.Errorf("%w: null supply request id", domain.ErrInvalidInput)
	}

	supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}
	if supplyRequest.BuyerID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	supplyOffers, err := uc.supplyOfferRepo.ListByRequest(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.SupplyOfferDTO, 0, len(supplyOffers))
	for i := range supplyOffers {
		// A rejected offer was already passed on. Re-surfacing it would let the
		// buyer pick the same supplier again on a request they already declined.
		// Matched offers stay: the buyer must keep seeing the offer they committed to.
		if supplyOffers[i].Status == domain.OfferRejected {
			continue
		}
		dtos = append(dtos, supplyOfferToDTO(&supplyOffers[i]))
	}

	return dtos, nil
}

func (uc *SupplyOfferUseCaseImpl) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyOfferDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if supplierID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	supplyOffers, err := uc.supplyOfferRepo.List(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.SupplyOfferDTO, 0, len(supplyOffers))
	for i := range supplyOffers {
		dtos = append(dtos, supplyOfferToDTO(&supplyOffers[i]))
	}

	return dtos, nil
}

func (uc *SupplyOfferUseCaseImpl) getOwnedActionableOffer(ctx context.Context, id uuid.UUID) (domain.SupplyOffer, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return domain.SupplyOffer{}, err
	}
	if !isFarmer(principal) {
		return domain.SupplyOffer{}, domain.ErrForbidden
	}
	if id == uuid.Nil {
		return domain.SupplyOffer{}, fmt.Errorf("%w: null supply offer id", domain.ErrInvalidInput)
	}

	supplyOffer, err := uc.supplyOfferRepo.GetByID(ctx, id)
	if err != nil {
		return domain.SupplyOffer{}, err
	}
	if supplyOffer.SupplierID != principal.UserID {
		return domain.SupplyOffer{}, domain.ErrForbidden
	}
	if !supplyOffer.IsActionable() {
		return domain.SupplyOffer{}, domain.ErrInvalidOfferStatus
	}

	return supplyOffer, nil
}

var _ primary.SupplyOfferUseCase = (*SupplyOfferUseCaseImpl)(nil)

func validateOfferAgainstRequest(supplyRequest domain.SupplyRequest, totalAmount float64) error {
	if supplyRequest.MultipleProviders && totalAmount < supplyRequest.MinAmountPerProvider {
		return fmt.Errorf("%w: offer total amount is below the minimum amount per provider", domain.ErrInvalidInput)
	}
	if totalAmount > supplyRequest.RemainingAmount() {
		return fmt.Errorf("%w: offer total amount exceeds the remaining amount of the supply request", domain.ErrInsufficientAmount)
	}
	return nil
}

func supplyOfferToDTO(supplyOffer *domain.SupplyOffer) *dto.SupplyOfferDTO {
	id := supplyOffer.ID
	supplierID := supplyOffer.SupplierID
	supplyRequestID := supplyOffer.SupplyRequest

	return &dto.SupplyOfferDTO{
		ID:                  &id,
		SupplierID:          &supplierID,
		SupplyRequest:       &supplyRequestID,
		TotalAmount:         supplyOffer.TotalAmount,
		AmountUnit:          supplyOffer.AmountUnit,
		PricePerUnit:        supplyOffer.PricePerUnit,
		Comments:            supplyOffer.Comments,
		ProposedDeliveryDay: supplyOffer.ProposedDeliveryDay,
		DeliveryAvailable:   supplyOffer.DeliveryAvailable,
		Status:              supplyOffer.Status,
		CreatedAt:           supplyOffer.CreatedAt,
		UpdatedAt:           supplyOffer.UpdatedAt,
	}
}
