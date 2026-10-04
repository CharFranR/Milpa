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
	if !isFarmer(principal) {
		return nil, domain.ErrForbidden
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
	if req.AllocationMethod != domain.AllocationManual {
		if req.AllocationMethod != domain.AllocationFirstCome {
			return nil, domain.ErrInvalidInput
		}
		liq.AllocationMethod = req.AllocationMethod
	}
	if req.ExpiresAt != nil {
		liq.SetExpiry(*req.ExpiresAt, now)
	}

	if err := uc.liquidationRepo.Save(ctx, liq); err != nil {
		return nil, fmt.Errorf("CreateLiquidation: %w", err)
	}

	return liquidationToDTO(liq), nil
}

// GetByID reads a liquidation the caller is allowed to see. A private one that
// belongs to somebody else is reported as not found rather than forbidden, so
// the response does not confirm that it exists.
func (uc *LiquidationUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.LiquidationDTO, error) {
	liq, err := uc.liquidationRepo.FindVisibleByID(ctx, id, liquidationViewer(ctx))
	if err != nil {
		return nil, err
	}

	return liquidationToDTO(liq), nil
}

func (uc *LiquidationUseCaseImpl) GetBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.LiquidationDTO, error) {
	liquidations, err := uc.liquidationRepo.FindBySupplier(ctx, supplierID, liquidationViewer(ctx))
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
	viewer := liquidationViewer(ctx)

	liquidations, err := uc.liquidationRepo.FindOpen(ctx, viewer)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.LiquidationDTO, 0, len(liquidations))
	for i := range liquidations {
		if !liquidationVisibleTo(&liquidations[i], viewer) {
			continue
		}
		dtos = append(dtos, liquidationToDTO(&liquidations[i]))
	}

	return dtos, nil
}

func liquidationVisibleTo(liq *domain.Liquidation, viewer port.LiquidationViewer) bool {
	if liq.SupplierID == viewer.ID {
		return true
	}
	switch liq.Visibility {
	case "public":
		return true
	case "wholesale":
		return viewer.SeesWholesale()
	case "wholesale_retail":
		return viewer.SeesWholesaleRetail()
	case "wholesale_corporate":
		return viewer.SeesWholesaleCorporate()
	default:
		return false
	}
}

// liquidationViewer resolves who is asking, for the visibility predicate.
func liquidationViewer(ctx context.Context) port.LiquidationViewer {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return port.LiquidationViewer{}
	}
	return port.LiquidationViewer{
		ID:   principal.UserID,
		Role: principal.Role,
	}
}

func (uc *LiquidationUseCaseImpl) UpdateLiquidation(ctx context.Context, id uuid.UUID, req dto.UpdateLiquidationRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	if !isFarmer(principal) {
		return domain.ErrForbidden
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
	if req.AllocationMethod != nil {
		if *req.AllocationMethod != domain.AllocationManual && *req.AllocationMethod != domain.AllocationFirstCome {
			return domain.ErrInvalidInput
		}
		liq.AllocationMethod = *req.AllocationMethod
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
	if !isFarmer(principal) {
		return domain.ErrForbidden
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

func (uc *LiquidationUseCaseImpl) ExpressInterest(ctx context.Context, liquidationID uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	if !isBuyer(principal) {
		return domain.ErrForbidden
	}

	liq, err := uc.liquidationRepo.FindVisibleByID(ctx, liquidationID, liquidationViewer(ctx))
	if err != nil {
		return err
	}
	if !liq.IsOpen() {
		return domain.ErrLiquidationNotOpen
	}

	exists, err := uc.liquidationRepo.InterestExists(ctx, liquidationID, principal.UserID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrInterestAlreadyExists
	}

	interest := domain.NewLiquidationInterest(liquidationID, principal.UserID, uc.timer.Now())
	return uc.liquidationRepo.SaveInterest(ctx, interest)
}

func (uc *LiquidationUseCaseImpl) ListInterests(ctx context.Context, liquidationID uuid.UUID) ([]*dto.LiquidationInterestDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	liq, err := uc.liquidationRepo.FindByID(ctx, liquidationID)
	if err != nil {
		return nil, err
	}
	if liq.SupplierID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	interests, err := uc.liquidationRepo.FindInterests(ctx, liquidationID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.LiquidationInterestDTO, 0, len(interests))
	for i := range interests {
		buyer, err := uc.userRepo.FindByID(ctx, interests[i].BuyerID)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, &dto.LiquidationInterestDTO{
			ID:            interests[i].ID,
			LiquidationID: interests[i].LiquidationID,
			BuyerID:       interests[i].BuyerID,
			BuyerName:     buyer.FullName(),
			CreatedAt:     interests[i].CreatedAt,
		})
	}

	return dtos, nil
}

func (uc *LiquidationUseCaseImpl) AssignLiquidation(ctx context.Context, liquidationID uuid.UUID, buyerID *uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	if !isFarmer(principal) {
		return domain.ErrForbidden
	}

	liq, err := uc.liquidationRepo.FindByID(ctx, liquidationID)
	if err != nil {
		return err
	}
	if liq.SupplierID != principal.UserID {
		return domain.ErrForbidden
	}
	if !liq.IsOpen() {
		return domain.ErrLiquidationCannotAssign
	}

	var chosen uuid.UUID
	if liq.AllocationMethod == domain.AllocationFirstCome {
		interests, err := uc.liquidationRepo.FindInterests(ctx, liquidationID)
		if err != nil {
			return err
		}
		earliest, ok := earliestInterest(interests)
		if !ok {
			return domain.ErrNoInterestToAssign
		}
		chosen = earliest.BuyerID
	} else {
		if buyerID == nil {
			return domain.ErrBuyerDidNotExpressInterest
		}
		exists, err := uc.liquidationRepo.InterestExists(ctx, liquidationID, *buyerID)
		if err != nil {
			return err
		}
		if !exists {
			return domain.ErrBuyerDidNotExpressInterest
		}
		chosen = *buyerID
	}

	if err := liq.Assign(chosen, uc.timer.Now()); err != nil {
		return err
	}

	return uc.liquidationRepo.Update(ctx, liq)
}

func earliestInterest(interests []domain.LiquidationInterest) (*domain.LiquidationInterest, bool) {
	var earliest *domain.LiquidationInterest
	for i := range interests {
		candidate := &interests[i]
		if earliest == nil ||
			candidate.CreatedAt.Before(earliest.CreatedAt) ||
			(candidate.CreatedAt.Equal(earliest.CreatedAt) && candidate.ID.String() < earliest.ID.String()) {
			earliest = candidate
		}
	}
	return earliest, earliest != nil
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
		AssignedBuyerID:  liq.AssignedBuyerID,
		CreatedAt:        liq.CreatedAt,
		UpdatedAt:        liq.UpdatedAt,
	}
}
