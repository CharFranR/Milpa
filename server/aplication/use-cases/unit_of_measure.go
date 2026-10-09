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

type UnitOfMeasureUseCaseImpl struct {
	unitRepo port.UnitOfMeasureRepository
}

func NewUnitOfMeasureUseCase(unitRepo port.UnitOfMeasureRepository) *UnitOfMeasureUseCaseImpl {
	return &UnitOfMeasureUseCaseImpl{unitRepo: unitRepo}
}

func (uc *UnitOfMeasureUseCaseImpl) List(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error) {
	units, err := uc.unitRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.UnitOfMeasureDTO, 0, len(units))
	for i := range units {
		dtos = append(dtos, unitToDTO(&units[i]))
	}

	return dtos, nil
}

func (uc *UnitOfMeasureUseCaseImpl) ListAll(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	units, err := uc.unitRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.UnitOfMeasureDTO, 0, len(units))
	for i := range units {
		dtos = append(dtos, unitToDTO(&units[i]))
	}

	return dtos, nil
}

func (uc *UnitOfMeasureUseCaseImpl) Create(ctx context.Context, req dto.CreateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	unit, err := domain.NewUnitOfMeasure(req.Code, req.Name)
	if err != nil {
		return nil, err
	}

	if err := uc.unitRepo.Save(ctx, unit); err != nil {
		return nil, err
	}

	return unitToDTO(unit), nil
}

func (uc *UnitOfMeasureUseCaseImpl) Update(ctx context.Context, id uuid.UUID, req dto.UpdateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	unit, err := uc.unitRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Code != nil {
		if err := unit.SetCode(*req.Code); err != nil {
			return nil, err
		}
	}
	if req.Name != nil {
		if err := unit.SetName(*req.Name); err != nil {
			return nil, err
		}
	}

	if err := uc.unitRepo.Save(ctx, unit); err != nil {
		return nil, err
	}

	return unitToDTO(unit), nil
}

func (uc *UnitOfMeasureUseCaseImpl) SetStatus(ctx context.Context, id uuid.UUID, req dto.UnitOfMeasureStatusRequest) (*dto.UnitOfMeasureDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	unit, err := uc.unitRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.IsActive {
		unit.Activate()
	} else {
		unit.Deactivate()
	}

	if err := uc.unitRepo.Save(ctx, unit); err != nil {
		return nil, err
	}

	return unitToDTO(unit), nil
}

var _ primary.UnitOfMeasureUseCase = (*UnitOfMeasureUseCaseImpl)(nil)

func unitToDTO(unit *domain.UnitOfMeasure) *dto.UnitOfMeasureDTO {
	return &dto.UnitOfMeasureDTO{
		ID:       unit.ID,
		Code:     unit.Code,
		Name:     unit.Name,
		IsActive: unit.IsActive,
	}
}
