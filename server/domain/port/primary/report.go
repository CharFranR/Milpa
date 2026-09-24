package primary

import (
	"context"

	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type ReportUseCase interface {
	Create(ctx context.Context, req dto.CreateReportRequest) (*dto.ReportResponse, error)
	List(ctx context.Context, status string, targetType string, page, pageSize int) (*dto.PaginatedReportsResponse, error)
	Resolve(ctx context.Context, id uuid.UUID, req dto.ResolveReportRequest) (*dto.ReportResponse, error)
}

type ModerationUseCase interface {
	SuspendUser(ctx context.Context, id uuid.UUID, req dto.SuspendUserRequest) error
	DeleteOffering(ctx context.Context, id uuid.UUID) error
	ListAuditLogs(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) (*dto.PaginatedAuditLogsResponse, error)
}
