package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type CategoryRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{pool: pool}
}

func (r *CategoryRepositoryImpl) FindAll(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, description FROM categories`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("category.FindAll: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description); err != nil {
			return nil, fmt.Errorf("category.FindAll: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category.FindAll: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `SELECT id, name, description FROM categories WHERE id = $1`

	var cat domain.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(&cat.ID, &cat.Name, &cat.Description)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("category.FindByID: %w", err)
	}

	return &cat, nil
}

var _ port.CategoryRepository = (*CategoryRepositoryImpl)(nil)
