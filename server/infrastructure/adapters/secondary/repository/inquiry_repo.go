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

type InquiryRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewInquiryRepository(pool *pgxpool.Pool) *InquiryRepositoryImpl {
	return &InquiryRepositoryImpl{pool: pool}
}

func (r *InquiryRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
	query := `
		SELECT id, user_id, offering_id, message, status, created_at
		FROM inquiries
		WHERE id = $1
	`

	var inquiry domain.Inquiry
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inquiry.ID, &inquiry.UserID, &inquiry.OfferingID, &inquiry.Message, &inquiry.Status, &inquiry.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("inquiry.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("inquiry.FindByID: %w", err)
	}

	return &inquiry, nil
}

func (r *InquiryRepositoryImpl) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error) {
	query := `
		SELECT id, user_id, offering_id, message, status, created_at
		FROM inquiries
		WHERE user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("inquiry.FindByUser: %w", err)
	}
	defer rows.Close()

	var inquiries []domain.Inquiry
	for rows.Next() {
		var inquiry domain.Inquiry
		if err := rows.Scan(
			&inquiry.ID, &inquiry.UserID, &inquiry.OfferingID, &inquiry.Message, &inquiry.Status, &inquiry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("inquiry.FindByUser: %w", err)
		}
		inquiries = append(inquiries, inquiry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inquiry.FindByUser: %w", err)
	}

	return inquiries, nil
}

func (r *InquiryRepositoryImpl) Save(ctx context.Context, inquiry *domain.Inquiry) error {
	query := `
		INSERT INTO inquiries (id, user_id, offering_id, message, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		inquiry.ID, inquiry.UserID, inquiry.OfferingID, inquiry.Message, inquiry.Status, inquiry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inquiry.Save: %w", err)
	}
	return nil
}

func (r *InquiryRepositoryImpl) Update(ctx context.Context, inquiry *domain.Inquiry) error {
	query := `UPDATE inquiries SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, inquiry.Status, inquiry.ID)
	if err != nil {
		return fmt.Errorf("inquiry.Update: %w", err)
	}
	return nil
}

var _ port.InquiryRepository = (*InquiryRepositoryImpl)(nil)
