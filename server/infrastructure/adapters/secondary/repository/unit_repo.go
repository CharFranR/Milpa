package repository

import (
	"context"
	"fmt"

	domain "milpa/domain/entities"
)

type UnitOfMeasureRepositoryImpl struct {
	pool DB
}

func NewUnitOfMeasureRepository(pool DB) *UnitOfMeasureRepositoryImpl {
	return &UnitOfMeasureRepositoryImpl{pool: pool}
}

func (r *UnitOfMeasureRepositoryImpl) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	const query = `SELECT id, code, name FROM units_of_measure WHERE is_active ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
	}
	defer rows.Close()

	units := make([]domain.UnitOfMeasure, 0)
	for rows.Next() {
		var unit domain.UnitOfMeasure
		if err := rows.Scan(&unit.ID, &unit.Code, &unit.Name); err != nil {
			return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
		}
		unit.IsActive = true
		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
	}

	return units, nil
}