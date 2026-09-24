package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

func TestReportUseCaseCreate(t *testing.T) {
	t.Run("happy path offering", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "offering",
			TargetID:   testTargetID.String(),
			Reason:     "Este producto parece fraudulento, no existe en mi zona",
		}

		result, err := uc.Create(principalCtx(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected report response, got nil")
		}
		if result.Status != "pending" {
			t.Errorf("expected status pending, got %s", result.Status)
		}
		if result.TargetType != "offering" {
			t.Errorf("expected target_type offering, got %s", result.TargetType)
		}
		if len(reportRepo.saved) != 1 {
			t.Errorf("expected 1 report saved, got %d", len(reportRepo.saved))
		}
		if len(auditRepo.saved) != 1 {
			t.Errorf("expected 1 audit log saved, got %d", len(auditRepo.saved))
		}
	})

	t.Run("happy path user", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "user",
			TargetID:   testOtherID.String(),
			Reason:     "Este usuario es un estafador, vende productos falsos",
		}

		result, err := uc.Create(principalCtx(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TargetType != "user" {
			t.Errorf("expected target_type user, got %s", result.TargetType)
		}
	})

	t.Run("self report", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "user",
			TargetID:   testUserID.String(),
			Reason:     "Quiero reportarme a mi mismo",
		}

		_, err := uc.Create(principalCtx(), req)
		if err == nil {
			t.Fatal("expected error for self report, got nil")
		}
		if err != domain.ErrSelfReport {
			t.Errorf("expected ErrSelfReport, got %v", err)
		}
	})

	t.Run("invalid target type", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "invalid",
			TargetID:   testTargetID.String(),
			Reason:     "Test reason that is long enough",
		}

		_, err := uc.Create(principalCtx(), req)
		if err == nil {
			t.Fatal("expected error for invalid target type, got nil")
		}
		if err != domain.ErrInvalidReportTargetType {
			t.Errorf("expected ErrInvalidReportTargetType, got %v", err)
		}
	})

	t.Run("reason too short", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "offering",
			TargetID:   testTargetID.String(),
			Reason:     "short",
		}

		_, err := uc.Create(principalCtx(), req)
		if err == nil {
			t.Fatal("expected error for short reason, got nil")
		}
	})

	t.Run("duplicate pending report", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.existsPending = func(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error) {
			return true, nil
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "offering",
			TargetID:   testTargetID.String(),
			Reason:     "Este producto parece fraudulento, no existe en mi zona",
		}

		_, err := uc.Create(principalCtx(), req)
		if err == nil {
			t.Fatal("expected error for duplicate pending report, got nil")
		}
		if err != domain.ErrReportAlreadyPending {
			t.Errorf("expected ErrReportAlreadyPending, got %v", err)
		}
	})

	t.Run("no auth", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.CreateReportRequest{
			TargetType: "offering",
			TargetID:   testTargetID.String(),
			Reason:     "Test reason that is long enough",
		}

		_, err := uc.Create(principalCtx(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestReportUseCaseList(t *testing.T) {
	t.Run("happy path admin", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		result, err := uc.List(adminCtx(), "", "", 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result.Total != 1 {
			t.Errorf("expected total 1, got %d", result.Total)
		}
		if len(result.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(result.Items))
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		_, err := uc.List(principalCtx(), "", "", 1, 20)
		if err == nil {
			t.Fatal("expected error for non-admin, got nil")
		}
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.findAll = func(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
			t.Error("FindAll must not be called for an unknown status")
			return nil, 0, nil
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		_, err := uc.List(adminCtx(), "bogus", "", 1, 20)
		if err == nil {
			t.Fatal("expected error for unknown status, got nil")
		}
		if err != domain.ErrInvalidReportStatus {
			t.Errorf("expected ErrInvalidReportStatus, got %v", err)
		}
	})

	t.Run("invalid target type", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.findAll = func(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
			t.Error("FindAll must not be called for an unknown target type")
			return nil, 0, nil
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		_, err := uc.List(adminCtx(), "", "bogus", 1, 20)
		if err == nil {
			t.Fatal("expected error for unknown target type, got nil")
		}
		if err != domain.ErrInvalidReportTargetType {
			t.Errorf("expected ErrInvalidReportTargetType, got %v", err)
		}
	})
}

func TestReportUseCaseResolve(t *testing.T) {
	t.Run("approve report offering", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "approve"}
		result, err := uc.Resolve(adminCtx(), testReportID, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "approved" {
			t.Errorf("expected status approved, got %s", result.Status)
		}
		if len(reportRepo.resolved) != 1 {
			t.Errorf("expected 1 report resolved, got %d", len(reportRepo.resolved))
		}
		if len(reportRepo.cascadeResolved) != 1 {
			t.Errorf("expected 1 sibling cascade after a successful resolve, got %d", len(reportRepo.cascadeResolved))
		}
	})

	t.Run("reject report", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "reject"}
		result, err := uc.Resolve(adminCtx(), testReportID, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "rejected" {
			t.Errorf("expected status rejected, got %s", result.Status)
		}
		if len(reportRepo.cascadeResolved) != 1 {
			t.Errorf("expected 1 sibling cascade after a successful resolve, got %d", len(reportRepo.cascadeResolved))
		}
	})

	t.Run("resolve failure skips sibling cascade", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.resolve = func(ctx context.Context, report *domain.Report) error {
			return errors.New("resolve failed")
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "approve"}
		_, err := uc.Resolve(adminCtx(), testReportID, req)
		if err == nil {
			t.Fatal("expected error when resolve fails, got nil")
		}
		if len(reportRepo.cascadeResolved) != 0 {
			t.Errorf("expected no sibling cascade after a failed resolve, got %d", len(reportRepo.cascadeResolved))
		}
	})

	t.Run("already resolved", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
			report := mustReport()
			report.Status = domain.ReportApproved
			return report, nil
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "approve"}
		_, err := uc.Resolve(adminCtx(), testReportID, req)
		if err == nil {
			t.Fatal("expected error for already resolved report, got nil")
		}
		if err != domain.ErrReportAlreadyResolved {
			t.Errorf("expected ErrReportAlreadyResolved, got %v", err)
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "invalid"}
		_, err := uc.Resolve(adminCtx(), testReportID, req)
		if err == nil {
			t.Fatal("expected error for invalid action, got nil")
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "approve"}
		_, err := uc.Resolve(principalCtx(), testReportID, req)
		if err == nil {
			t.Fatal("expected error for non-admin, got nil")
		}
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}
