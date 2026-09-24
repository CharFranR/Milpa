package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type ReportUseCaseImpl struct {
	reportRepo   port.ReportRepository
	auditRepo    port.AuditLogRepository
	userRepo     port.UserRepository
	offeringRepo port.OfferingRepository
	timer        port.TimeProvider
}

func NewReportUseCase(
	reportRepo port.ReportRepository,
	auditRepo port.AuditLogRepository,
	userRepo port.UserRepository,
	offeringRepo port.OfferingRepository,
	timer port.TimeProvider,
) *ReportUseCaseImpl {
	return &ReportUseCaseImpl{
		reportRepo:   reportRepo,
		auditRepo:    auditRepo,
		userRepo:     userRepo,
		offeringRepo: offeringRepo,
		timer:        timer,
	}
}

func (uc *ReportUseCaseImpl) Create(ctx context.Context, req dto.CreateReportRequest) (*dto.ReportResponse, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	targetType := domain.ReportTargetType(req.TargetType)
	if targetType != domain.ReportTargetOffering && targetType != domain.ReportTargetUser {
		return nil, domain.ErrInvalidReportTargetType
	}

	if strings.TrimSpace(req.Reason) == "" {
		return nil, domain.ErrReasonRequired
	}
	if len(req.Reason) < 10 || len(req.Reason) > 500 {
		return nil, fmt.Errorf("reason must be between 10 and 500 characters")
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	if targetType == domain.ReportTargetUser && targetID == principal.UserID {
		return nil, domain.ErrSelfReport
	}

	exists, err := uc.reportRepo.ExistsPendingByTarget(ctx, principal.UserID, targetType, targetID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrReportAlreadyPending
	}

	now := uc.timer.Now()

	report, err := domain.NewReport(principal.UserID, targetType, targetID, req.Reason, now)
	if err != nil {
		return nil, err
	}

	metadata, _ := json.Marshal(map[string]string{"report_id": report.ID.String()})
	auditLog := domain.NewAuditLog(principal.UserID, domain.AuditActionReportCreated, "report", report.ID, metadata, now)

	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, err
	}
	if err := uc.auditRepo.Save(ctx, auditLog); err != nil {
		return nil, err
	}

	return uc.buildResponse(ctx, report)
}

func (uc *ReportUseCaseImpl) List(ctx context.Context, status string, targetType string, page, pageSize int) (*dto.PaginatedReportsResponse, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if status != "" {
		switch domain.ReportStatus(status) {
		case domain.ReportPending, domain.ReportApproved, domain.ReportRejected:
		default:
			return nil, domain.ErrInvalidReportStatus
		}
	}
	if targetType != "" && targetType != string(domain.ReportTargetOffering) && targetType != string(domain.ReportTargetUser) {
		return nil, domain.ErrInvalidReportTargetType
	}

	reports, total, err := uc.reportRepo.FindAll(ctx, status, targetType, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ReportResponse, len(reports))
	for i := range reports {
		resp, err := uc.buildResponse(ctx, &reports[i])
		if err != nil {
			return nil, err
		}
		items[i] = *resp
	}

	return &dto.PaginatedReportsResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (uc *ReportUseCaseImpl) Resolve(ctx context.Context, id uuid.UUID, req dto.ResolveReportRequest) (*dto.ReportResponse, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	action := strings.ToLower(req.Action)
	if action != "approve" && action != "reject" {
		return nil, fmt.Errorf("action must be 'approve' or 'reject'")
	}

	report, err := uc.reportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if report.IsResolved() {
		return nil, domain.ErrReportAlreadyResolved
	}

	now := uc.timer.Now()

	if action == "approve" {
		report.Approve(principal.UserID, now)

		if report.TargetType == domain.ReportTargetUser {
			user, err := uc.userRepo.FindByID(ctx, report.TargetID)
			if err != nil {
				return nil, err
			}
			user.Suspend(now)
			if err := uc.userRepo.Update(ctx, user); err != nil {
				return nil, err
			}
			suspendMeta, _ := json.Marshal(map[string]string{
				"report_id": report.ID.String(),
				"user_id":   report.TargetID.String(),
			})
			suspendLog := domain.NewAuditLog(principal.UserID, domain.AuditActionUserSuspended, "user", report.TargetID, suspendMeta, now)
			_ = uc.auditRepo.Save(ctx, suspendLog)
		} else {
			if err := uc.offeringRepo.Delete(ctx, report.TargetID); err != nil {
				return nil, err
			}
			deleteMeta, _ := json.Marshal(map[string]string{
				"report_id":   report.ID.String(),
				"offering_id": report.TargetID.String(),
			})
			deleteLog := domain.NewAuditLog(principal.UserID, domain.AuditActionOfferingDeleted, "offering", report.TargetID, deleteMeta, now)
			_ = uc.auditRepo.Save(ctx, deleteLog)
		}

		approveMeta, _ := json.Marshal(map[string]string{"report_id": report.ID.String()})
		approveLog := domain.NewAuditLog(principal.UserID, domain.AuditActionReportApproved, "report", report.ID, approveMeta, now)
		_ = uc.auditRepo.Save(ctx, approveLog)
	} else {
		report.Reject(principal.UserID, now)

		rejectMeta, _ := json.Marshal(map[string]string{"report_id": report.ID.String()})
		rejectLog := domain.NewAuditLog(principal.UserID, domain.AuditActionReportRejected, "report", report.ID, rejectMeta, now)
		_ = uc.auditRepo.Save(ctx, rejectLog)
	}

	if err := uc.reportRepo.Resolve(ctx, report); err != nil {
		return nil, err
	}

	return uc.buildResponse(ctx, report)
}

func (uc *ReportUseCaseImpl) buildResponse(ctx context.Context, report *domain.Report) (*dto.ReportResponse, error) {
	reporter, err := uc.userRepo.FindByID(ctx, report.ReporterID)
	if err != nil {
		return nil, err
	}

	targetName := ""
	if report.TargetType == domain.ReportTargetOffering {
		offering, err := uc.offeringRepo.FindByID(ctx, report.TargetID)
		if err == nil {
			targetName = offering.Name
		}
	} else {
		user, err := uc.userRepo.FindByID(ctx, report.TargetID)
		if err == nil {
			targetName = user.Email
		}
	}

	return &dto.ReportResponse{
		ID: report.ID,
		Reporter: dto.UserBrief{
			ID:    reporter.ID,
			Name:  reporter.FullName(),
			Email: reporter.Email,
		},
		TargetType: string(report.TargetType),
		Target: dto.TargetBrief{
			ID:   report.TargetID,
			Name: targetName,
		},
		Reason:     report.Reason,
		Status:     string(report.Status),
		ResolvedBy: report.ResolvedBy,
		ResolvedAt: report.ResolvedAt,
		CreatedAt:  report.CreatedAt,
	}, nil
}

var _ primary.ReportUseCase = (*ReportUseCaseImpl)(nil)

type ModerationUseCaseImpl struct {
	userRepo     port.UserRepository
	offeringRepo port.OfferingRepository
	auditRepo    port.AuditLogRepository
	offeringUC   primary.OfferingUseCase
	timer        port.TimeProvider
}

func NewModerationUseCase(
	userRepo port.UserRepository,
	offeringRepo port.OfferingRepository,
	auditRepo port.AuditLogRepository,
	offeringUC primary.OfferingUseCase,
	timer port.TimeProvider,
) *ModerationUseCaseImpl {
	return &ModerationUseCaseImpl{
		userRepo:     userRepo,
		offeringRepo: offeringRepo,
		auditRepo:    auditRepo,
		offeringUC:   offeringUC,
		timer:        timer,
	}
}

func (uc *ModerationUseCaseImpl) SuspendUser(ctx context.Context, id uuid.UUID, req dto.SuspendUserRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	if principal.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	action := strings.ToLower(req.Action)
	if action != "suspend" && action != "reactivate" {
		return fmt.Errorf("action must be 'suspend' or 'reactivate'")
	}

	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if action == "suspend" {
		if user.ID == principal.UserID {
			return domain.ErrCannotSuspendSelf
		}
		if user.IsAdmin() {
			return domain.ErrCannotSuspendAdmin
		}
		user.Suspend(uc.timer.Now())

		meta, _ := json.Marshal(map[string]string{"user_id": id.String()})
		log := domain.NewAuditLog(principal.UserID, domain.AuditActionUserSuspended, "user", id, meta, uc.timer.Now())
		_ = uc.auditRepo.Save(ctx, log)
	} else {
		user.Reactivate(uc.timer.Now())

		meta, _ := json.Marshal(map[string]string{"user_id": id.String()})
		log := domain.NewAuditLog(principal.UserID, domain.AuditActionUserReactivated, "user", id, meta, uc.timer.Now())
		_ = uc.auditRepo.Save(ctx, log)
	}

	return uc.userRepo.Update(ctx, user)
}

func (uc *ModerationUseCaseImpl) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	if principal.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if err := uc.offeringUC.DeleteOffering(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return err
	}

	meta, _ := json.Marshal(map[string]string{"offering_id": id.String()})
	log := domain.NewAuditLog(principal.UserID, domain.AuditActionOfferingDeleted, "offering", id, meta, uc.timer.Now())
	_ = uc.auditRepo.Save(ctx, log)

	return nil
}

func (uc *ModerationUseCaseImpl) ListAuditLogs(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) (*dto.PaginatedAuditLogsResponse, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := uc.auditRepo.FindAll(ctx, action, actorID, targetType, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AuditLogResponse, len(logs))
	for i, l := range logs {
		var meta any
		if l.Metadata != nil {
			_ = json.Unmarshal(l.Metadata, &meta)
		}
		items[i] = dto.AuditLogResponse{
			ID:         l.ID,
			ActorID:    l.ActorID,
			Action:     l.Action,
			TargetType: l.TargetType,
			TargetID:   l.TargetID,
			Metadata:   meta,
			CreatedAt:  l.CreatedAt,
		}
	}

	return &dto.PaginatedAuditLogsResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

var _ primary.ModerationUseCase = (*ModerationUseCaseImpl)(nil)
