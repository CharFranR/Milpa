package repository

import (
	"context"
	"fmt"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type AdminStatsRepositoryImpl struct {
	pool DB
}

func NewAdminStatsRepository(pool DB) *AdminStatsRepositoryImpl {
	return &AdminStatsRepositoryImpl{pool: pool}
}

func (r *AdminStatsRepositoryImpl) Counts(ctx context.Context) (domain.AdminStats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM users WHERE suspended_at IS NOT NULL),
			(SELECT COUNT(*) FROM offerings WHERE ` + cataloguePredicate + `),
			(SELECT COUNT(*) FROM liquidations WHERE status = 'open'),
			(SELECT COUNT(*) FROM reports WHERE status = 'pending')
	`

	var stats domain.AdminStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.SuspendedUsers,
		&stats.ActiveOfferings,
		&stats.OpenLiquidations,
		&stats.PendingReports,
	)
	if err != nil {
		return domain.AdminStats{}, fmt.Errorf("adminStats.Counts: %w", err)
	}

	return stats, nil
}

var _ port.AdminStatsRepository = (*AdminStatsRepositoryImpl)(nil)
