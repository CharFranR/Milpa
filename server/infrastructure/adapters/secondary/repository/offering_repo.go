package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

const offeringColumns = `id, user_id, type, name, description, price, image_url,
	variety, unit_of_measure_id, quantity_available, expires_at, is_active,
	category_id, company_id, latitude, longitude, created_at, updated_at`

const cataloguePredicate = `is_active AND (expires_at IS NULL OR expires_at > NOW())`

type OfferingRepositoryImpl struct {
	pool DB
}

func NewOfferingRepository(pool DB) *OfferingRepositoryImpl {
	return &OfferingRepositoryImpl{pool: pool}
}

func scanOffering(scan func(dest ...any) error) (domain.Offering, error) {
	var offering domain.Offering
	err := scan(
		&offering.ID, &offering.UserID, &offering.Type, &offering.Name, &offering.Description,
		&offering.Price, &offering.ImageURL, &offering.Variety, &offering.UnitOfMeasureID,
		&offering.QuantityAvailable, &offering.ExpiresAt, &offering.IsActive, &offering.CategoryID,
		&offering.CompanyID, &offering.Latitude, &offering.Longitude, &offering.CreatedAt, &offering.UpdatedAt,
	)
	return offering, err
}

func (r *OfferingRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
	query := `SELECT ` + offeringColumns + `
		FROM offerings
		WHERE id = $1
	`

	offering, err := scanOffering(r.pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("offering.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("offering.FindByID: %w", err)
	}

	return &offering, nil
}

func (r *OfferingRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Offering, error) {
	query := `SELECT ` + offeringColumns + `
		FROM offerings
		WHERE user_id = $1 AND ` + cataloguePredicate + `
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("offering.FindByUser: %w", err)
	}
	defer rows.Close()

	var offerings []domain.Offering
	for rows.Next() {
		offering, err := scanOffering(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("offering.FindByUser: %w", err)
		}
		offerings = append(offerings, offering)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("offering.FindByUser: %w", err)
	}

	return offerings, nil
}

func (r *OfferingRepositoryImpl) DeactivateExpired(ctx context.Context, now time.Time) ([]domain.Offering, error) {
	query := `
		UPDATE offerings
		SET is_active = false, updated_at = $2
		WHERE is_active AND expires_at <= $1
		RETURNING ` + offeringColumns + `
	`

	rows, err := r.pool.Query(ctx, query, now, now)
	if err != nil {
		return nil, fmt.Errorf("offering.DeactivateExpired: %w", err)
	}
	defer rows.Close()

	var offerings []domain.Offering
	for rows.Next() {
		offering, err := scanOffering(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("offering.DeactivateExpired: %w", err)
		}
		offerings = append(offerings, offering)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("offering.DeactivateExpired: %w", err)
	}

	return offerings, nil
}

func (r *OfferingRepositoryImpl) Save(ctx context.Context, offering *domain.Offering) error {
	query := `
		INSERT INTO offerings (id, user_id, type, name, description, price, image_url,
			variety, unit_of_measure_id, quantity_available, expires_at, is_active,
			category_id, company_id, latitude, longitude, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := r.pool.Exec(ctx, query,
		offering.ID, offering.UserID, offering.Type, offering.Name, offering.Description,
		offering.Price, offering.ImageURL, offering.Variety, offering.UnitOfMeasureID,
		offering.QuantityAvailable, offering.ExpiresAt, offering.IsActive, offering.CategoryID,
		offering.CompanyID, offering.Latitude, offering.Longitude, offering.CreatedAt, offering.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("offering.Save: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("offering.Save: %w", err)
	}
	return nil
}

func (r *OfferingRepositoryImpl) Update(ctx context.Context, offering *domain.Offering) error {
	query := `
		UPDATE offerings
		SET user_id = $1, type = $2, name = $3, description = $4, price = $5, image_url = $6,
			variety = $7, unit_of_measure_id = $8, quantity_available = $9, expires_at = $10,
			is_active = $11, category_id = $12, company_id = $13, latitude = $14, longitude = $15,
			updated_at = $16
		WHERE id = $17
	`
	_, err := r.pool.Exec(ctx, query,
		offering.UserID, offering.Type, offering.Name, offering.Description, offering.Price,
		offering.ImageURL, offering.Variety, offering.UnitOfMeasureID, offering.QuantityAvailable,
		offering.ExpiresAt, offering.IsActive, offering.CategoryID, offering.CompanyID,
		offering.Latitude, offering.Longitude, offering.UpdatedAt, offering.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("offering.Update: %w", domain.ErrDuplicate)
		}
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
