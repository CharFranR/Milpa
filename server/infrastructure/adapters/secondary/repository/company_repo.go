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

type CompanyRepositoryImpl struct {
	pool DB
}

func NewCompanyRepository(pool DB) *CompanyRepositoryImpl {
	return &CompanyRepositoryImpl{pool: pool}
}

func (r *CompanyRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	query := `
		SELECT c.id, c.name, c.owner_id, c.description, c.phone_number, c.email, c.website, c.verified, c.created_at, c.updated_at,
		       a.id, a.department, a.municipality, a.address_line, a.latitude, a.longitude
		FROM companies c
		LEFT JOIN addresses a ON c.address_id = a.id
		WHERE c.id = $1
	`

	var company domain.Company
	var ownerID uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&company.ID, &company.Name, &ownerID, &company.Description, &company.PhoneNumber,
		&company.Email, &company.Website, &company.Verified, &company.CreatedAt, &company.UpdatedAt,
		&company.Address.ID, &company.Address.Department, &company.Address.Municipality, &company.Address.AddressLine,
		&company.Address.Latitude, &company.Address.Longitude,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("company.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("company.FindByID: %w", err)
	}

	company.Owner = domain.User{ID: ownerID}

	categories, err := r.fetchCategories(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("company.FindByID: %w", err)
	}
	company.Category = categories

	return &company, nil
}

func (r *CompanyRepositoryImpl) FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
	query := `
		SELECT c.id, c.name, c.owner_id, c.description, c.phone_number, c.email, c.website, c.verified, c.created_at, c.updated_at,
		       a.id, a.department, a.municipality, a.address_line, a.latitude, a.longitude
		FROM companies c
		LEFT JOIN addresses a ON c.address_id = a.id
		WHERE c.owner_id = $1
	`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("company.FindByOwner: %w", err)
	}
	defer rows.Close()

	var companies []domain.Company
	for rows.Next() {
		var company domain.Company
		var scannedOwnerID uuid.UUID

		if err := rows.Scan(
			&company.ID, &company.Name, &scannedOwnerID, &company.Description, &company.PhoneNumber,
			&company.Email, &company.Website, &company.Verified, &company.CreatedAt, &company.UpdatedAt,
			&company.Address.ID, &company.Address.Department, &company.Address.Municipality, &company.Address.AddressLine,
			&company.Address.Latitude, &company.Address.Longitude,
		); err != nil {
			return nil, fmt.Errorf("company.FindByOwner: %w", err)
		}

		company.Owner = domain.User{ID: scannedOwnerID}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("company.FindByOwner: %w", err)
	}

	// Fetch categories for each company found
	for i := range companies {
		categories, err := r.fetchCategories(ctx, companies[i].ID)
		if err != nil {
			return nil, fmt.Errorf("company.FindByOwner: %w", err)
		}
		companies[i].Category = categories
	}

	return companies, nil
}

func (r *CompanyRepositoryImpl) Save(ctx context.Context, company *domain.Company) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("company.Save: %w", err)
	}
	defer tx.Rollback(ctx)

	var addressID uuid.UUID

	if company.Address.Department != "" {
		query := `
			INSERT INTO addresses (id, department, municipality, address_line, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		err = tx.QueryRow(ctx, query, uuid.New(), company.Address.Department, company.Address.Municipality,
			company.Address.AddressLine, company.Address.Latitude, company.Address.Longitude).Scan(&addressID)

		if err != nil {
			return fmt.Errorf("company.Save: %w", err)
		}
	}

	query := `
		INSERT INTO companies (id, name, owner_id, address_id, description, phone_number, email, website, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, query,
		company.ID, company.Name, company.Owner.ID, nullUUID(addressID),
		company.Description, company.PhoneNumber, company.Email, company.Website,
		company.Verified, company.CreatedAt, company.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("company.Save: insert company: %w", err)
	}

	for _, category := range company.Category {
		_, err = tx.Exec(ctx, "INSERT INTO company_categories (company_id, category_id) VALUES ($1, $2)", company.ID, category.ID)
		if err != nil {
			return fmt.Errorf("company.Save: insert category: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *CompanyRepositoryImpl) Update(ctx context.Context, company *domain.Company) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("company.Update: %w", err)
	}
	defer tx.Rollback(ctx)

	if company.Address.ID != uuid.Nil {
		query := `
			UPDATE addresses
			SET department = $1, municipality = $2, address_line = $3, latitude = $4, longitude = $5
			WHERE id = $6
		`
		_, err = tx.Exec(ctx, query,
			company.Address.Department, company.Address.Municipality, company.Address.AddressLine,
			company.Address.Latitude, company.Address.Longitude, company.Address.ID,
		)
		if err != nil {
			return fmt.Errorf("company.Update: address update: %w", err)
		}
	}

	query := `
		UPDATE companies
		SET name = $1, description = $2, phone_number = $3, email = $4, website = $5, verified = $6,
		    updated_at = $7, address_id = $8
		WHERE id = $9
	`
	_, err = tx.Exec(ctx, query,
		company.Name, company.Description, company.PhoneNumber, company.Email, company.Website,
		company.Verified, company.UpdatedAt, nullUUID(company.Address.ID), company.ID,
	)
	if err != nil {
		return fmt.Errorf("company.Update: update company: %w", err)
	}

	// Sync categories: remove existing and insert current set
	_, err = tx.Exec(ctx, "DELETE FROM company_categories WHERE company_id = $1", company.ID)
	if err != nil {
		return fmt.Errorf("company.Update: delete categories: %w", err)
	}

	for _, category := range company.Category {
		_, err = tx.Exec(ctx, "INSERT INTO company_categories (company_id, category_id) VALUES ($1, $2)", company.ID, category.ID)
		if err != nil {
			return fmt.Errorf("company.Update: insert category: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// fetchCategories retrieves all categories associated with a company.
func (r *CompanyRepositoryImpl) fetchCategories(ctx context.Context, companyID uuid.UUID) ([]domain.Category, error) {
	query := `
		SELECT c.id, c.name, c.description
		FROM categories c
		JOIN company_categories cc ON c.id = cc.category_id
		WHERE cc.company_id = $1
	`

	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

var _ port.CompanyRepository = (*CompanyRepositoryImpl)(nil)
