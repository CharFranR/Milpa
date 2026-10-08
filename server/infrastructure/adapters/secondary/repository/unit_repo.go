package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

const unitOfMeasureColumns = `id, code, name, is_active`

type UnitOfMeasureRepositoryImpl struct {
	pool DB
}

func NewUnitOfMeasureRepository(pool DB) *UnitOfMeasureRepositoryImpl {
	return &UnitOfMeasureRepositoryImpl{pool: pool}
}

func scanUnit(scan func(dest ...any) error) (domain.UnitOfMeasure, error) {
	var unit domain.UnitOfMeasure
	err := scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	return unit, err
}

func (r *UnitOfMeasureRepositoryImpl) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	query := `SELECT ` + unitOfMeasureColumns + ` FROM units_of_measure WHERE is_active ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
	}
	defer rows.Close()

	units := make([]domain.UnitOfMeasure, 0)
	for rows.Next() {
		unit, err := scanUnit(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
		}
		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unitOfMeasure.List: %w", err)
	}

	return units, nil
}

func (r *UnitOfMeasureRepositoryImpl) FindAll(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	query := `SELECT ` + unitOfMeasureColumns + ` FROM units_of_measure ORDER BY name, id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unitOfMeasure.FindAll: %w", err)
	}
	defer rows.Close()

	units := make([]domain.UnitOfMeasure, 0)
	for rows.Next() {
		unit, err := scanUnit(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("unitOfMeasure.FindAll: %w", err)
		}
		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unitOfMeasure.FindAll: %w", err)
	}

	return units, nil
}

func (r *UnitOfMeasureRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	query := `SELECT ` + unitOfMeasureColumns + ` FROM units_of_measure WHERE id = $1`

	unit, err := scanUnit(r.pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("unitOfMeasure.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("unitOfMeasure.FindByID: %w", err)
	}

	return &unit, nil
}

func (r *UnitOfMeasureRepositoryImpl) Save(ctx context.Context, unit *domain.UnitOfMeasure) error {
	query := `
		INSERT INTO units_of_measure (id, code, name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			name = EXCLUDED.name,
			is_active = EXCLUDED.is_active
	`
	_, err := r.pool.Exec(ctx, query, unit.ID, unit.Code, unit.Name, unit.IsActive)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("unitOfMeasure.Save: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("unitOfMeasure.Save: %w", err)
	}
	return nil
}

var _ port.UnitOfMeasureRepository = (*UnitOfMeasureRepositoryImpl)(nil)
