package usecases

import (
	"context"

	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"
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
		dtos = append(dtos, &dto.UnitOfMeasureDTO{
			ID:   units[i].ID,
			Code: units[i].Code,
			Name: units[i].Name,
		})
	}

	return dtos, nil
}