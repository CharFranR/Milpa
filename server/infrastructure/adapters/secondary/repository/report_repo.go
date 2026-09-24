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

type ReportRepositoryImpl struct {
	pool DB
}

func NewReportRepository(pool DB) *ReportRepositoryImpl {
	return &ReportRepositoryImpl{pool: pool}
}

func (r *ReportRepositoryImpl) Save(ctx context.Context, report *domain.Report) error {
	query := `
		INSERT INTO reports (id, reporter_id, target_type, target_id, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		report.ID, report.ReporterID, report.TargetType, report.TargetID,
		report.Reason, report.Status, report.CreatedAt, report.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("report.Save: %w", err)
	}
	return nil
}

func (r *ReportRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	query := `
		SELECT id, reporter_id, target_type, target_id, reason, status, resolved_by, resolved_at, created_at, updated_at
		FROM reports
		WHERE id = $1
	`

	var report domain.Report
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&report.ID, &report.ReporterID, &report.TargetType, &report.TargetID,
		&report.Reason, &report.Status, &report.ResolvedBy, &report.ResolvedAt,
		&report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("report.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("report.FindByID: %w", err)
	}

	return &report, nil
}

func (r *ReportRepositoryImpl) FindAll(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
	countQuery := `SELECT COUNT(*) FROM reports WHERE 1=1`
	query := `
		SELECT id, reporter_id, target_type, target_id, reason, status, resolved_by, resolved_at, created_at, updated_at
		FROM reports WHERE 1=1
	`

	args := []any{}
	argIdx := 1

	if status != "" {
		countQuery += fmt.Sprintf(` AND status = $%d`, argIdx)
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, status)
		argIdx++
	}
	if targetType != "" {
		countQuery += fmt.Sprintf(` AND target_type = $%d`, argIdx)
		query += fmt.Sprintf(` AND target_type = $%d`, argIdx)
		args = append(args, targetType)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("report.FindAll count: %w", err)
	}

	offset := (page - 1) * pageSize
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("report.FindAll: %w", err)
	}
	defer rows.Close()

	var reports []domain.Report
	for rows.Next() {
		var report domain.Report
		if err := rows.Scan(
			&report.ID, &report.ReporterID, &report.TargetType, &report.TargetID,
			&report.Reason, &report.Status, &report.ResolvedBy, &report.ResolvedAt,
			&report.CreatedAt, &report.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("report.FindAll scan: %w", err)
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("report.FindAll rows: %w", err)
	}

	return reports, total, nil
}

func (r *ReportRepositoryImpl) Resolve(ctx context.Context, report *domain.Report) error {
	query := `
		UPDATE reports
		SET status = $1, resolved_by = $2, resolved_at = $3, updated_at = $4
		WHERE id = $5 AND status = 'pending'
	`
	tag, err := r.pool.Exec(ctx, query,
		report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.ID,
	)
	if err != nil {
		return fmt.Errorf("report.Resolve: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return r.resolveMissError(ctx, report.ID)
	}
	return nil
}

func (r *ReportRepositoryImpl) resolveMissError(ctx context.Context, id uuid.UUID) error {
	query := `SELECT status FROM reports WHERE id = $1`
	var status domain.ReportStatus
	err := r.pool.QueryRow(ctx, query, id).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("report.Resolve: %w", domain.ErrNotFound)
		}
		return fmt.Errorf("report.Resolve: %w", err)
	}
	return fmt.Errorf("report.Resolve: %w", domain.ErrReportAlreadyResolved)
}

func (r *ReportRepositoryImpl) ResolvePendingByTarget(ctx context.Context, report *domain.Report) error {
	query := `
		UPDATE reports
		SET status = $1, resolved_by = $2, resolved_at = $3, updated_at = $4
		WHERE target_type = $5 AND target_id = $6 AND status = 'pending'
	`
	_, err := r.pool.Exec(ctx, query,
		report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt,
		report.TargetType, report.TargetID,
	)
	if err != nil {
		return fmt.Errorf("report.ResolvePendingByTarget: %w", err)
	}
	return nil
}

func (r *ReportRepositoryImpl) ExistsPendingByTarget(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reports
			WHERE reporter_id = $1 AND target_type = $2 AND target_id = $3 AND status = 'pending'
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, reporterID, targetType, targetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("report.ExistsPendingByTarget: %w", err)
	}
	return exists, nil
}

var _ port.ReportRepository = (*ReportRepositoryImpl)(nil)
