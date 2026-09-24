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

type AuditLogRepositoryImpl struct {
	pool DB
}

func NewAuditLogRepository(pool DB) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{pool: pool}
}

func (r *AuditLogRepositoryImpl) Save(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor_id, action, target_type, target_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		log.ID, log.ActorID, log.Action, log.TargetType, log.TargetID,
		log.Metadata, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("audit_log.Save: %w", err)
	}
	return nil
}

func (r *AuditLogRepositoryImpl) FindAll(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error) {
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	query := `
		SELECT id, actor_id, action, target_type, target_id, metadata, created_at
		FROM audit_logs WHERE 1=1
	`

	args := []any{}
	argIdx := 1

	if action != "" {
		countQuery += fmt.Sprintf(` AND action = $%d`, argIdx)
		query += fmt.Sprintf(` AND action = $%d`, argIdx)
		args = append(args, action)
		argIdx++
	}
	if actorID != "" {
		countQuery += fmt.Sprintf(` AND actor_id = $%d`, argIdx)
		query += fmt.Sprintf(` AND actor_id = $%d`, argIdx)
		args = append(args, actorID)
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
		return nil, 0, fmt.Errorf("audit_log.FindAll count: %w", err)
	}

	offset := (page - 1) * pageSize
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_log.FindAll: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		if err := rows.Scan(
			&log.ID, &log.ActorID, &log.Action, &log.TargetType, &log.TargetID,
			&log.Metadata, &log.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("audit_log.FindAll scan: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit_log.FindAll rows: %w", err)
	}

	return logs, total, nil
}

// Enforce append-only: no Update or Delete methods

func (r *AuditLogRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	query := `
		SELECT id, actor_id, action, target_type, target_id, metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`

	var log domain.AuditLog
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.ActorID, &log.Action, &log.TargetType, &log.TargetID,
		&log.Metadata, &log.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("audit_log.FindByID: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("audit_log.FindByID: %w", err)
	}

	return &log, nil
}

var _ port.AuditLogRepository = (*AuditLogRepositoryImpl)(nil)
