package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type SupplyRequestUseCaseImpl struct {
	supplyRequestRepo port.SupplyRequestRepository
	offerRepo         port.SupplyOfferRepository
	matchRepo         port.MatchRepository
	timer             port.TimeProvider
}

func NewSupplyRequestUseCase(supplyRequestRepo port.SupplyRequestRepository, offerRepo port.SupplyOfferRepository, matchRepo port.MatchRepository, timer port.TimeProvider) *SupplyRequestUseCaseImpl {
	return &SupplyRequestUseCaseImpl{
		supplyRequestRepo: supplyRequestRepo,
		offerRepo:         offerRepo,
		matchRepo:         matchRepo,
		timer:             timer,
	}
}

func (uc *SupplyRequestUseCaseImpl) Create(ctx context.Context, req dto.SupplyRequestDTO) (*dto.SupplyRequestDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateSupplyRequestContent(req.ProductName, req.TotalAmount); err != nil {
		return nil, err
	}
	if err := validateSupplyRequestDeadlines(req.RequestDeadline, req.DeliveryDeadline); err != nil {
		return nil, err
	}

	supplyRequest := domain.NewSupplyRequest(
		principal.UserID, req.ProductName, req.TotalAmount, req.AmountUnit, req.NumberOfUnits,
		req.AmountPerUnit, req.UnitOfMeasure, req.Address, req.RequestDeadline, req.DeliveryDeadline,
		req.Description, req.MultipleProviders,
	)
	supplyRequest.MinAmountPerProvider = req.MinAmountPerProvider

	if err := uc.supplyRequestRepo.Create(ctx, supplyRequest); err != nil {
		return nil, err
	}

	return supplyRequestToDTO(supplyRequest), nil
}

func (uc *SupplyRequestUseCaseImpl) List(ctx context.Context) ([]*dto.SupplyRequestDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	supplyRequests, err := uc.supplyRequestRepo.List(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.SupplyRequestDTO, len(supplyRequests))
	for i := range supplyRequests {
		dtos[i] = supplyRequestToDTO(&supplyRequests[i])
	}

	return dtos, nil
}

func (uc *SupplyRequestUseCaseImpl) ListAvailable(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyRequestDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if supplierID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	supplyRequests, err := uc.supplyRequestRepo.ListOpen(ctx)
	if err != nil {
		return nil, err
	}

	offers, err := uc.offerRepo.List(ctx, supplierID)
	if err != nil {
		return nil, err
	}
	offeredRequests := make(map[uuid.UUID]struct{}, len(offers))
	for i := range offers {
		offeredRequests[offers[i].SupplyRequest] = struct{}{}
	}

	dtos := make([]*dto.SupplyRequestDTO, 0, len(supplyRequests))
	for i := range supplyRequests {
		if supplyRequests[i].BuyerID == supplierID {
			continue
		}
		if supplyRequests[i].ActualAmount <= 0 {
			continue
		}
		if _, ok := offeredRequests[supplyRequests[i].ID]; ok {
			continue
		}
		dtos = append(dtos, supplyRequestToDTO(&supplyRequests[i]))
	}

	return dtos, nil
}

func (uc *SupplyRequestUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyRequestDTO, error) {
	if _, err := auth.RequirePrincipal(ctx); err != nil {
		return nil, err
	}
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: null supply request id", domain.ErrInvalidInput)
	}

	supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return supplyRequestToDTO(&supplyRequest), nil
}

func (uc *SupplyRequestUseCaseImpl) Update(ctx context.Context, id uuid.UUID, req dto.SupplyGeneralUpdateDTO) error {
	supplyRequest, err := uc.getOwnedSupplyRequest(ctx, id)
	if err != nil {
		return err
	}
	if !supplyRequest.IsOpen() {
		return domain.ErrInvalidRequestStatus
	}
	if err := validateSupplyRequestContent(req.ProductName, req.TotalAmount); err != nil {
		return err
	}
	if err := validateSupplyAmounts(req.TotalAmount, req.ActualAmount); err != nil {
		return err
	}
	if err := validateSupplyRequestDeadlines(req.RequestDeadline, req.DeliveryDeadline); err != nil {
		return err
	}
	if err := uc.validateAgainstMatchedAmount(ctx, id, req.TotalAmount, req.ActualAmount); err != nil {
		return err
	}

	supplyRequest.ProductName = req.ProductName
	supplyRequest.TotalAmount = req.TotalAmount
	supplyRequest.ActualAmount = req.ActualAmount
	supplyRequest.AmountUnit = req.AmountUnit
	supplyRequest.NumberOfUnits = req.NumberOfUnits
	supplyRequest.AmountPerUnit = req.AmountPerUnit
	supplyRequest.UnitOfMeasure = req.UnitOfMeasure
	supplyRequest.Address = req.Address
	supplyRequest.RequestDeadline = req.RequestDeadline
	supplyRequest.DeliveryDeadline = req.DeliveryDeadline
	supplyRequest.Description = req.Description
	supplyRequest.MultipleProviders = req.MultipleProviders
	supplyRequest.MinAmountPerProvider = req.MinAmountPerProvider
	supplyRequest.UpdatedAt = uc.timer.Now()

	return uc.supplyRequestRepo.Update(ctx, supplyRequest)
}

func (uc *SupplyRequestUseCaseImpl) UpdateAmounts(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateAmountsDTO) error {
	supplyRequest, err := uc.getOwnedSupplyRequest(ctx, id)
	if err != nil {
		return err
	}
	if !supplyRequest.IsOpen() {
		return domain.ErrInvalidRequestStatus
	}
	if err := validateSupplyAmounts(req.TotalAmount, req.ActualAmount); err != nil {
		return err
	}
	if err := uc.validateAgainstMatchedAmount(ctx, id, req.TotalAmount, req.ActualAmount); err != nil {
		return err
	}

	supplyRequest.TotalAmount = req.TotalAmount
	supplyRequest.ActualAmount = req.ActualAmount
	supplyRequest.AmountUnit = req.AmountUnit
	supplyRequest.AmountPerUnit = req.AmountPerUnit
	supplyRequest.UnitOfMeasure = req.UnitOfMeasure
	supplyRequest.MultipleProviders = req.MultipleProviders
	supplyRequest.MinAmountPerProvider = req.MinAmountPerProvider
	supplyRequest.UpdatedAt = uc.timer.Now()

	return uc.supplyRequestRepo.Update(ctx, supplyRequest)
}

func (uc *SupplyRequestUseCaseImpl) UpdateDeadlines(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateTimeDTO) error {
	supplyRequest, err := uc.getOwnedSupplyRequest(ctx, id)
	if err != nil {
		return err
	}
	if !supplyRequest.IsOpen() {
		return domain.ErrInvalidRequestStatus
	}
	if err := validateSupplyRequestDeadlines(req.RequestDeadline, req.DeliveryDeadline); err != nil {
		return err
	}

	supplyRequest.RequestDeadline = req.RequestDeadline
	supplyRequest.DeliveryDeadline = req.DeliveryDeadline
	supplyRequest.UpdatedAt = uc.timer.Now()

	return uc.supplyRequestRepo.Update(ctx, supplyRequest)
}

func (uc *SupplyRequestUseCaseImpl) Cancel(ctx context.Context, id uuid.UUID) error {
	supplyRequest, err := uc.getOwnedSupplyRequest(ctx, id)
	if err != nil {
		return err
	}
	if !supplyRequest.IsOpen() {
		return domain.ErrInvalidRequestStatus
	}

	hasActiveMatch, err := uc.matchRepo.ExistsActiveByRequest(ctx, id)
	if err != nil {
		return err
	}
	if hasActiveMatch {
		return primary.ErrActiveMatch
	}

	if err := supplyRequest.Cancel(); err != nil {
		return err
	}

	return uc.supplyRequestRepo.Update(ctx, supplyRequest)
}

func (uc *SupplyRequestUseCaseImpl) Expire(ctx context.Context, id uuid.UUID) error {
	supplyRequest, err := uc.getOwnedSupplyRequest(ctx, id)
	if err != nil {
		return err
	}
	if !supplyRequest.IsOpen() {
		return domain.ErrInvalidRequestStatus
	}

	hasActiveMatch, err := uc.matchRepo.ExistsActiveByRequest(ctx, id)
	if err != nil {
		return err
	}
	if hasActiveMatch {
		return primary.ErrActiveMatch
	}

	if err := supplyRequest.Expire(); err != nil {
		return err
	}

	return uc.supplyRequestRepo.Update(ctx, supplyRequest)
}

func (uc *SupplyRequestUseCaseImpl) getOwnedSupplyRequest(ctx context.Context, id uuid.UUID) (*domain.SupplyRequest, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: null supply request id", domain.ErrInvalidInput)
	}

	supplyRequest, err := uc.supplyRequestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if supplyRequest.BuyerID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	return &supplyRequest, nil
}

func (uc *SupplyRequestUseCaseImpl) validateAgainstMatchedAmount(ctx context.Context, id uuid.UUID, totalAmount, actualAmount float64) error {
	matches, err := uc.matchRepo.ListActiveByRequest(ctx, id)
	if err != nil {
		return err
	}

	var matched float64
	for i := range matches {
		matched += matches[i].MatchedAmount
	}

	if totalAmount < matched {
		return fmt.Errorf("%w: total amount cannot be lower than the already matched amount", domain.ErrInsufficientAmount)
	}
	if actualAmount > totalAmount-matched {
		return fmt.Errorf("%w: actual amount cannot include already matched quantity", domain.ErrInsufficientAmount)
	}

	return nil
}

var _ primary.SupplyRequestUseCase = (*SupplyRequestUseCaseImpl)(nil)

func validateSupplyRequestContent(productName string, totalAmount float64) error {
	if productName == "" {
		return fmt.Errorf("%w: product name is required", domain.ErrInvalidInput)
	}
	if totalAmount <= 0 {
		return fmt.Errorf("%w: total amount must be greater than zero", domain.ErrInvalidInput)
	}
	return nil
}

func validateSupplyAmounts(totalAmount, actualAmount float64) error {
	if totalAmount <= 0 {
		return fmt.Errorf("%w: total amount must be greater than zero", domain.ErrInvalidInput)
	}
	if actualAmount < 0 || actualAmount > totalAmount {
		return fmt.Errorf("%w: actual amount must be between zero and total amount", domain.ErrInvalidInput)
	}
	return nil
}

func validateSupplyRequestDeadlines(requestDeadline, deliveryDeadline time.Time) error {
	if !requestDeadline.IsZero() && !deliveryDeadline.IsZero() && requestDeadline.After(deliveryDeadline) {
		return fmt.Errorf("%w: request deadline must not be after delivery deadline", domain.ErrInvalidInput)
	}
	return nil
}

func supplyRequestToDTO(supplyRequest *domain.SupplyRequest) *dto.SupplyRequestDTO {
	id := supplyRequest.ID
	buyerID := supplyRequest.BuyerID

	return &dto.SupplyRequestDTO{
		ID:                   &id,
		BuyerID:              &buyerID,
		ProductName:          supplyRequest.ProductName,
		TotalAmount:          supplyRequest.TotalAmount,
		ActualAmount:         supplyRequest.ActualAmount,
		AmountUnit:           supplyRequest.AmountUnit,
		NumberOfUnits:        supplyRequest.NumberOfUnits,
		AmountPerUnit:        supplyRequest.AmountPerUnit,
		UnitOfMeasure:        supplyRequest.UnitOfMeasure,
		Address:              supplyRequest.Address,
		RequestDeadline:      supplyRequest.RequestDeadline,
		DeliveryDeadline:     supplyRequest.DeliveryDeadline,
		Description:          supplyRequest.Description,
		MultipleProviders:    supplyRequest.MultipleProviders,
		MinAmountPerProvider: supplyRequest.MinAmountPerProvider,
		Status:               supplyRequest.Status,
		CreatedAt:            supplyRequest.CreatedAt,
		UpdatedAt:            supplyRequest.UpdatedAt,
	}
}
