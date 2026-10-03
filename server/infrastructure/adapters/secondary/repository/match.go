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

type MatchRepositoryImpl struct {
	db Querier
}

func NewMatchRepository(pool DB) *MatchRepositoryImpl {
	return &MatchRepositoryImpl{db: pool}
}

func scanMatch(scan func(dest ...any) error) (domain.Match, error) {
	var match domain.Match

	err := scan(
		&match.ID, &match.SupplyOffer, &match.SupplyRequest, &match.Status, &match.MatchedAmount, &match.AmountUnit,
		&match.CreatedAt, &match.UpdatedAt,
	)
	if err != nil {
		return domain.Match{}, err
	}

	return match, nil
}

func (r *MatchRepositoryImpl) Create(ctx context.Context, match *domain.Match) error {
	query := `
		INSERT INTO matches (id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		match.ID, match.SupplyOffer, match.SupplyRequest, match.Status, match.MatchedAmount, match.AmountUnit,
		match.CreatedAt, match.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("match.Create: %w", domain.ErrDuplicate)
		}
		return fmt.Errorf("match.Create: %w", err)
	}
	return nil
}

func (r *MatchRepositoryImpl) ListByOffer(ctx context.Context, supplyOfferID uuid.UUID) ([]domain.Match, error) {
	query := `
		SELECT id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at
		FROM matches
		WHERE supply_offer_id = $1
	`

	rows, err := r.db.Query(ctx, query, supplyOfferID)
	if err != nil {
		return nil, fmt.Errorf("match.ListByOffer: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		match, err := scanMatch(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("match.ListByOffer: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("match.ListByOffer: %w", err)
	}

	return matches, nil
}

func (r *MatchRepositoryImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	query := `
		SELECT id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at
		FROM matches
		WHERE supply_request_id = $1
	`

	rows, err := r.db.Query(ctx, query, supplyRequestID)
	if err != nil {
		return nil, fmt.Errorf("match.ListByRequest: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		match, err := scanMatch(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("match.ListByRequest: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("match.ListByRequest: %w", err)
	}

	return matches, nil
}

func (r *MatchRepositoryImpl) ListActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error) {
	query := `
		SELECT id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at
		FROM matches
		WHERE supply_request_id = $1 AND status = $2
	`

	rows, err := r.db.Query(ctx, query, supplyRequestID, domain.MatchActive)
	if err != nil {
		return nil, fmt.Errorf("match.ListActiveByRequest: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		match, err := scanMatch(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("match.ListActiveByRequest: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("match.ListActiveByRequest: %w", err)
	}

	return matches, nil
}

func (r *MatchRepositoryImpl) ListActiveBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Match, error) {
	query := `
		SELECT m.id, m.supply_offer_id, m.supply_request_id, m.status, m.matched_amount, m.amount_unit, m.created_at, m.updated_at
		FROM matches m
		JOIN supply_offers o ON m.supply_offer_id = o.id
		WHERE o.supplier_id = $1 AND m.status = $2
	`

	rows, err := r.db.Query(ctx, query, supplierID, domain.MatchActive)
	if err != nil {
		return nil, fmt.Errorf("match.ListActiveBySupplier: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		match, err := scanMatch(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("match.ListActiveBySupplier: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("match.ListActiveBySupplier: %w", err)
	}

	return matches, nil
}

func (r *MatchRepositoryImpl) ExistsActiveByRequest(ctx context.Context, supplyRequestID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM matches WHERE supply_request_id = $1 AND status = $2)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, supplyRequestID, domain.MatchActive).Scan(&exists); err != nil {
		return false, fmt.Errorf("match.ExistsActiveByRequest: %w", err)
	}

	return exists, nil
}

func (r *MatchRepositoryImpl) ExistsActiveByOffer(ctx context.Context, supplyOfferID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM matches WHERE supply_offer_id = $1 AND status = $2)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, supplyOfferID, domain.MatchActive).Scan(&exists); err != nil {
		return false, fmt.Errorf("match.ExistsActiveByOffer: %w", err)
	}

	return exists, nil
}

func (r *MatchRepositoryImpl) GetByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at
		FROM matches
		WHERE id = $1
	`

	match, err := scanMatch(r.db.QueryRow(ctx, query, matchID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("match.GetByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("match.GetByID: %w", err)
	}

	return &match, nil
}

// LockByIDForUpdate reads the match and holds a row lock on it until the
// enclosing transaction ends. It is the third lock of the global order:
// SupplyRequest -> SupplyOffer -> Match.
func (r *MatchRepositoryImpl) LockByIDForUpdate(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at
		FROM matches
		WHERE id = $1
		FOR UPDATE
	`

	match, err := scanMatch(r.db.QueryRow(ctx, query, matchID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("match.LockByIDForUpdate: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("match.LockByIDForUpdate: %w", err)
	}

	return &match, nil
}

func (r *MatchRepositoryImpl) Update(ctx context.Context, match *domain.Match) error {
	query := `
		UPDATE matches
		SET status = $1, matched_amount = $2, amount_unit = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, match.Status, match.MatchedAmount, match.AmountUnit, match.UpdatedAt, match.ID)
	if err != nil {
		return fmt.Errorf("match.Update: %w", err)
	}
	return nil
}

func (r *MatchRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM matches WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("match.Delete: %w", err)
	}
	return nil
}

var (
	_ port.MatchRepository   = (*MatchRepositoryImpl)(nil)
	_ port.TxMatchRepository = (*MatchRepositoryImpl)(nil)
)
