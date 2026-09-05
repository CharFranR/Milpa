package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type ReviewRepositoryImpl struct {
	pool DB
}

func NewReviewRepository(pool DB) *ReviewRepositoryImpl {
	return &ReviewRepositoryImpl{pool: pool}
}

func (r *ReviewRepositoryImpl) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
	query := `
		SELECT id, user_id, company_id, rating, comment, created_at
		FROM reviews
		WHERE company_id = $1
	`

	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("review.FindByCompany: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID, &review.UserID, &review.CompanyID, &review.Rating, &review.Comment, &review.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("review.FindByCompany: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review.FindByCompany: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepositoryImpl) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Review, error) {
	query := `
		SELECT id, user_id, company_id, rating, comment, created_at
		FROM reviews
		WHERE user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("review.FindByUser: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID, &review.UserID, &review.CompanyID, &review.Rating, &review.Comment, &review.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("review.FindByUser: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review.FindByUser: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepositoryImpl) Save(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, company_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		review.ID, review.UserID, review.CompanyID, review.Rating, review.Comment, review.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("review.Save: %w", err)
	}
	return nil
}

var _ port.ReviewRepository = (*ReviewRepositoryImpl)(nil)
