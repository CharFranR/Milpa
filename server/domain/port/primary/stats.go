package primary

import (
	"context"

	"milpa/aplication/dto"
)

type AdminStatsUseCase interface {
	Get(ctx context.Context) (*dto.AdminStatsDTO, error)
}
