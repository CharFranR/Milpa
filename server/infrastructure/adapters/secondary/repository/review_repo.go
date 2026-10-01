package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

const reviewColumns = `id, author_id, target_type, target_id, company_id, rating, comment, created_at`

func scanReview(scan func(dest ...any) error) (domain.Review, error) {
	var review domain.Review
	var companyID *uuid.UUID

	err := scan(
		&review.ID, &review.AuthorID, &review.TargetType, &review.TargetID, &companyID,
		&review.Rating, &review.Comment, &review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}

	if companyID != nil {
		review.CompanyID = *companyID
	}

	return review, nil
}

type ReviewRepositoryImpl struct {
	pool DB
}

func NewReviewRepository(pool DB) *ReviewRepositoryImpl {
	return &ReviewRepositoryImpl{pool: pool}
}

func (r *ReviewRepositoryImpl) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
	query := `
		SELECT ` + reviewColumns + `
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
		review, err := scanReview(rows.Scan)
		if err != nil {
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
		SELECT ` + reviewColumns + `
		FROM reviews
		WHERE author_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("review.FindByUser: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		review, err := scanReview(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("review.FindByUser: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review.FindByUser: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepositoryImpl) AverageRating(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) (float64, int, error) {
	query := `
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM reviews
		WHERE target_type = $1 AND target_id = $2
	`

	var average float64
	var count int
	if err := r.pool.QueryRow(ctx, query, targetType, targetID).Scan(&average, &count); err != nil {
		return 0, 0, fmt.Errorf("review.AverageRating: %w", err)
	}

	return average, count, nil
}

func (r *ReviewRepositoryImpl) Save(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, author_id, target_type, target_id, company_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		review.ID, review.AuthorID, review.TargetType, review.TargetID, nullUUID(review.CompanyID),
		review.Rating, review.Comment, review.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("review.Save: %w", err)
	}
	return nil
}

var _ port.ReviewRepository = (*ReviewRepositoryImpl)(nil)
