package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type LiquidationUseCaseImpl struct {
	liquidationRepo port.LiquidationRepository
	userRepo        port.UserRepository
	timer           port.TimeProvider
}

func NewLiquidationUseCase(liquidationRepo port.LiquidationRepository, userRepo port.UserRepository, timer port.TimeProvider) *LiquidationUseCaseImpl {
	return &LiquidationUseCaseImpl{
		liquidationRepo: liquidationRepo,
		userRepo:        userRepo,
		timer:           timer,
	}
}

func (uc *LiquidationUseCaseImpl) CreateLiquidation(ctx context.Context, req dto.CreateLiquidationRequest) (*dto.LiquidationDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	// Verify user exists
	user, err := uc.userRepo.FindByID(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	now := uc.timer.Now()

	liq, err := domain.NewLiquidation(
		user.ID,
		req.ProductName,
		req.Quantity,
		req.UnitOfMeasure,
		req.TotalPrice,
		req.UnitPrice,
		now,
	)
	if err != nil {
		return nil, err
	}

	if req.DeliveryTime != "" {
		liq.DeliveryTime = req.DeliveryTime
	}
	if req.LocationID != uuid.Nil {
		liq.LocationID = req.LocationID
	}
	if req.Visibility != "" {
		if err := liq.UpdateVisibility(req.Visibility, now); err != nil {
			return nil, err
		}
	}
	if req.ExpiresAt != nil {
		liq.SetExpiry(*req.ExpiresAt, now)
	}

	if err := uc.liquidationRepo.Save(ctx, liq); err != nil {
		return nil, fmt.Errorf("CreateLiquidation: %w", err)
	}

	return liquidationToDTO(liq), nil
}

func (uc *LiquidationUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.LiquidationDTO, error) {
	liq, err := uc.liquidationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return liquidationToDTO(liq), nil
}

func (uc *LiquidationUseCaseImpl) GetBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.LiquidationDTO, error) {
	liquidations, err := uc.liquidationRepo.FindBySupplier(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.LiquidationDTO, len(liquidations))
	for i := range liquidations {
		dtos[i] = liquidationToDTO(&liquidations[i])
	}

	return dtos, nil
}

func (uc *LiquidationUseCaseImpl) GetOpen(ctx context.Context) ([]*dto.LiquidationDTO, error) {
	liquidations, err := uc.liquidationRepo.FindOpen(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.LiquidationDTO, len(liquidations))
	for i := range liquidations {
		dtos[i] = liquidationToDTO(&liquidations[i])
	}

	return dtos, nil
}

func (uc *LiquidationUseCaseImpl) UpdateLiquidation(ctx context.Context, id uuid.UUID, req dto.UpdateLiquidationRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	liq, err := uc.liquidationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Only the supplier can update their liquidation
	if liq.SupplierID != principal.UserID {
		return domain.ErrForbidden
	}

	if !liq.IsOpen() {
		return domain.ErrLiquidationNotOpen
	}

	now := uc.timer.Now()

	if req.ProductName != nil {
		liq.ProductName = *req.ProductName
	}
	if req.Quantity != nil {
		liq.Quantity = *req.Quantity
	}
	if req.UnitOfMeasure != nil {
		liq.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.TotalPrice != nil {
		liq.TotalPrice = *req.TotalPrice
	}
	if req.UnitPrice != nil {
		liq.UnitPrice = *req.UnitPrice
	}
	if req.DeliveryTime != nil {
		liq.UpdateDeliveryTime(*req.DeliveryTime, now)
	}
	if req.LocationID != nil {
		liq.LocationID = *req.LocationID
	}
	if req.Visibility != nil {
		if err := liq.UpdateVisibility(*req.Visibility, now); err != nil {
			return err
		}
	}
	if req.ExpiresAt != nil {
		liq.SetExpiry(*req.ExpiresAt, now)
	}

	liq.Touch(now)

	return uc.liquidationRepo.Update(ctx, liq)
}

func (uc *LiquidationUseCaseImpl) DeleteLiquidation(ctx context.Context, id uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	liq, err := uc.liquidationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Only the supplier can delete their liquidation
	if liq.SupplierID != principal.UserID {
		return domain.ErrForbidden
	}

	return uc.liquidationRepo.Delete(ctx, id)
}

var _ primary.LiquidationUseCase = (*LiquidationUseCaseImpl)(nil)

func liquidationToDTO(liq *domain.Liquidation) *dto.LiquidationDTO {
	return &dto.LiquidationDTO{
		ID:               liq.ID,
		SupplierID:       liq.SupplierID,
		ProductName:      liq.ProductName,
		Quantity:         liq.Quantity,
		UnitOfMeasure:    liq.UnitOfMeasure,
		TotalPrice:       liq.TotalPrice,
		UnitPrice:        liq.UnitPrice,
		DeliveryTime:     liq.DeliveryTime,
		LocationID:       liq.LocationID,
		Visibility:       liq.Visibility,
		AllocationMethod: liq.AllocationMethod,
		Status:           liq.Status,
		ClosedAt:         liq.ClosedAt,
		ExpiresAt:        liq.ExpiresAt,
		CreatedAt:        liq.CreatedAt,
		UpdatedAt:        liq.UpdatedAt,
	}
}
