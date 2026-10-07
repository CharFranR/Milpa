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

const categoryColumns = `id, name, description, main_category, is_active, default_unit_of_measure_id, default_expiry_days`

type CategoryRepositoryImpl struct {
	pool DB
}

func NewCategoryRepository(pool DB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{pool: pool}
}

func scanCategory(scan func(dest ...any) error) (domain.Category, error) {
	var category domain.Category
	err := scan(
		&category.ID, &category.Name, &category.Description,
		&category.MainCategory, &category.IsActive, &category.DefaultUnitOfMeasureID,
		&category.DefaultExpiryDays,
	)
	return category, err
}

func (r *CategoryRepositoryImpl) FindAll(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT ` + categoryColumns + ` FROM categories`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("category.FindAll: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		category, err := scanCategory(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("category.FindAll: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category.FindAll: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `SELECT ` + categoryColumns + ` FROM categories WHERE id = $1`

	category, err := scanCategory(r.pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("category.FindByID: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepositoryImpl) Save(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, description, main_category, is_active, default_unit_of_measure_id, default_expiry_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			main_category = EXCLUDED.main_category,
			is_active = EXCLUDED.is_active,
			default_unit_of_measure_id = EXCLUDED.default_unit_of_measure_id,
			default_expiry_days = EXCLUDED.default_expiry_days
	`
	_, err := r.pool.Exec(ctx, query,
		category.ID, category.Name, category.Description,
		category.MainCategory, category.IsActive, category.DefaultUnitOfMeasureID,
		category.DefaultExpiryDays,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("category.Save: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("category.Save: %w", err)
	}
	return nil
}

var _ port.CategoryRepository = (*CategoryRepositoryImpl)(nil)
