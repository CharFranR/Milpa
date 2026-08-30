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

type OfferingRepositoryImpl struct {
	pool DB
}

func NewOfferingRepository(pool DB) *OfferingRepositoryImpl {
	return &OfferingRepositoryImpl{pool: pool}
}

func (r *OfferingRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
	query := `
		SELECT id, company_id, type, name, description, price, image_url, created_at, updated_at
		FROM offerings
		WHERE id = $1
	`

	var offering domain.Offering
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&offering.ID, &offering.CompanyID, &offering.Type, &offering.Name, &offering.Description,
		&offering.Price, &offering.ImageURL, &offering.CreatedAt, &offering.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("offering.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("offering.FindByID: %w", err)
	}

	return &offering, nil
}

func (r *OfferingRepositoryImpl) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error) {
	query := `
		SELECT id, company_id, type, name, description, price, image_url, created_at, updated_at
		FROM offerings
		WHERE company_id = $1
	`

	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("offering.FindByCompany: %w", err)
	}
	defer rows.Close()

	var offerings []domain.Offering
	for rows.Next() {
		var offering domain.Offering
		if err := rows.Scan(
			&offering.ID, &offering.CompanyID, &offering.Type, &offering.Name, &offering.Description,
			&offering.Price, &offering.ImageURL, &offering.CreatedAt, &offering.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("offering.FindByCompany: %w", err)
		}
		offerings = append(offerings, offering)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("offering.FindByCompany: %w", err)
	}

	return offerings, nil
}

func (r *OfferingRepositoryImpl) Save(ctx context.Context, offering *domain.Offering) error {
	query := `
		INSERT INTO offerings (id, company_id, type, name, description, price, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		offering.ID, offering.CompanyID, offering.Type, offering.Name, offering.Description,
		offering.Price, offering.ImageURL, offering.CreatedAt, offering.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("offering.Save: %w", err)
	}
	return nil
}

func (r *OfferingRepositoryImpl) Update(ctx context.Context, offering *domain.Offering) error {
	query := `
		UPDATE offerings
		SET company_id = $1, type = $2, name = $3, description = $4, price = $5, image_url = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.pool.Exec(ctx, query,
		offering.CompanyID, offering.Type, offering.Name, offering.Description,
		offering.Price, offering.ImageURL, offering.UpdatedAt, offering.ID,
	)
	if err != nil {
		return fmt.Errorf("offering.Update: %w", err)
	}
	return nil
}

func (r *OfferingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM offerings WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("offering.Delete: %w", err)
	}
	return nil
}

var _ port.OfferingRepository = (*OfferingRepositoryImpl)(nil)
