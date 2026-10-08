package usecases

import (
	"context"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type AdminStatsUseCaseImpl struct {
	statsRepo port.AdminStatsRepository
}

func NewAdminStatsUseCase(statsRepo port.AdminStatsRepository) *AdminStatsUseCaseImpl {
	return &AdminStatsUseCaseImpl{statsRepo: statsRepo}
}

func (uc *AdminStatsUseCaseImpl) Get(ctx context.Context) (*dto.AdminStatsDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	stats, err := uc.statsRepo.Counts(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.AdminStatsDTO{
		TotalUsers:       stats.TotalUsers,
		SuspendedUsers:   stats.SuspendedUsers,
		ActiveOfferings:  stats.ActiveOfferings,
		OpenLiquidations: stats.OpenLiquidations,
		PendingReports:   stats.PendingReports,
	}, nil
}

var _ primary.AdminStatsUseCase = (*AdminStatsUseCaseImpl)(nil)
