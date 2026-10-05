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

		result, err := uc.List(reportAdminCtx(), "", "", 1, 20)
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

		_, err := uc.List(reportAdminCtx(), "bogus", "", 1, 20)
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

		_, err := uc.List(reportAdminCtx(), "", "bogus", 1, 20)
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
		result, err := uc.Resolve(reportAdminCtx(), testReportID, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "approved" {
			t.Errorf("expected status approved, got %s", result.Status)
		}
		if len(reportRepo.resolved) != 1 {
			t.Errorf("expected 1 report resolved, got %d", len(reportRepo.resolved))
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
		result, err := uc.Resolve(reportAdminCtx(), testReportID, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "rejected" {
			t.Errorf("expected status rejected, got %s", result.Status)
		}
		if len(reportRepo.resolved) != 1 {
			t.Errorf("expected 1 report resolved, got %d", len(reportRepo.resolved))
		}
	})

	t.Run("resolve failure stops before response building", func(t *testing.T) {
		reportRepo := newFakeReportRepo()
		reportRepo.resolve = func(ctx context.Context, report *domain.Report) error {
			return errors.New("resolve failed")
		}
		auditRepo := newFakeAuditLogRepo()
		userRepo := newFakeUserRepo()
		userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			t.Error("FindByID must not be called after a failed resolve")
			return nil, nil
		}
		offeringRepo := newFakeOfferingRepo()
		timer := newFakeTimer()

		uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, timer)

		req := dto.ResolveReportRequest{Action: "approve"}
		result, err := uc.Resolve(reportAdminCtx(), testReportID, req)
		if err == nil {
			t.Fatal("expected error when resolve fails, got nil")
		}
		if err.Error() != "resolve failed" {
			t.Errorf("expected the resolve error to propagate, got %v", err)
		}
		if result != nil {
			t.Errorf("expected no response after a failed resolve, got %+v", result)
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
		_, err := uc.Resolve(reportAdminCtx(), testReportID, req)
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
		_, err := uc.Resolve(reportAdminCtx(), testReportID, req)
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

// errAuditWrite stands in for an audit_logs insert that fails — a dropped
// connection, a constraint, a full disk. Every moderation action is required to
// be logged ("el sistema registrará todas las operaciones de moderación para
// auditoría"), so a write that fails must abort the action rather than commit it
// unlogged.
var errAuditWrite = errors.New("audit log insert failed")

// failingAuditRepo is newFakeAuditLogRepo with every Save rejecting, so the
// only way a moderation use case can succeed is by ignoring the failure.
func failingAuditRepo() *fakeAuditLogRepo {
	repo := newFakeAuditLogRepo()
	repo.save = func(ctx context.Context, log *domain.AuditLog) error {
		return errAuditWrite
	}
	return repo
}

// TestReportUseCaseResolveFailsClosedOnAuditWrite is the RF-16 guarantee for
// report resolution.
//
// Report creation already checked the audit error; every one of the four writes
// on the resolution path discarded it with `_ = uc.auditRepo.Save(...)`, so an
// admin could approve or reject a report, and the platform would be left holding
// a moderation action with no record of who took it or why.
func TestReportUseCaseResolveFailsClosedOnAuditWrite(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"approve", "reject"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			reportRepo := newFakeReportRepo()
			auditRepo := failingAuditRepo()
			userRepo := newFakeUserRepo()
			offeringRepo := newFakeOfferingRepo()

			uc := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, newFakeTimer())

			result, err := uc.Resolve(reportAdminCtx(), testReportID, dto.ResolveReportRequest{Action: action})

			if !errors.Is(err, errAuditWrite) {
				t.Fatalf("expected the audit write error to propagate, got %v", err)
			}
			if result != nil {
				t.Errorf("expected no response after a failed audit write, got %+v", result)
			}
			// The guarantee: the report is not handed to the repository, so the
			// stored row keeps its pending status and the queue still shows it.
			if len(reportRepo.resolved) != 0 {
				t.Errorf("report.Resolve called %d times, want 0: a failed audit write must leave the report pending", len(reportRepo.resolved))
			}
			if len(auditRepo.saved) != 0 {
				t.Errorf("audit logs stored = %d, want 0", len(auditRepo.saved))
			}
		})
	}
}

// TestModerationUseCaseSuspendUserFailsClosedOnAuditWrite is the RF-16
// guarantee for the moderation endpoint.
//
// SuspendUser writes the audit log BEFORE it persists the user, so failing the
// audit write is what keeps an unsuspended user from becoming suspended without
// a trace. This is the assertion that matters: not merely that an error is
// returned, but that the user is still active afterwards.
func TestModerationUseCaseSuspendUserFailsClosedOnAuditWrite(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"suspend", "reactivate"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			offeringRepo := newFakeOfferingRepo()
			auditRepo := failingAuditRepo()
			uc := usecases.NewModerationUseCase(userRepo, offeringRepo, auditRepo, newFakeTimer())

			err := uc.SuspendUser(reportAdminCtx(), testUserID, dto.SuspendUserRequest{Action: action})

			if !errors.Is(err, errAuditWrite) {
				t.Fatalf("expected the audit write error to propagate, got %v", err)
			}
			if len(userRepo.updated) != 0 {
				t.Errorf("user.Update called %d times, want 0: a failed audit write must leave the user unsuspended", len(userRepo.updated))
			}
		})
	}
}

// TestModerationSuspendUserPersistsWhenAuditSucceeds is the control for the
// test above. Without it, a use case that never wrote the user at all would
// satisfy the fail-closed assertion while the feature was quietly broken.
func TestModerationSuspendUserPersistsWhenAuditSucceeds(t *testing.T) {
	t.Parallel()

	userRepo := newFakeUserRepo()
	offeringRepo := newFakeOfferingRepo()
	auditRepo := newFakeAuditLogRepo()
	uc := usecases.NewModerationUseCase(userRepo, offeringRepo, auditRepo, newFakeTimer())

	if err := uc.SuspendUser(reportAdminCtx(), testUserID, dto.SuspendUserRequest{Action: "suspend"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.saved) != 1 {
		t.Fatalf("audit logs stored = %d, want 1", len(auditRepo.saved))
	}
	if len(userRepo.updated) != 1 {
		t.Fatalf("users updated = %d, want 1", len(userRepo.updated))
	}
	if !userRepo.updated[0].IsSuspended() {
		t.Error("the user must be suspended once the audit write has succeeded")
	}
}

// TestModerationUseCaseDeleteOfferingFailsClosedOnAuditWrite is the RF-16
// guarantee for offering deletion: the action reports failure instead of
// returning a success the audit trail does not support.
func TestModerationUseCaseDeleteOfferingFailsClosedOnAuditWrite(t *testing.T) {
	t.Parallel()

	userRepo := newFakeUserRepo()
	offeringRepo := newFakeOfferingRepo()
	auditRepo := failingAuditRepo()
	uc := usecases.NewModerationUseCase(userRepo, offeringRepo, auditRepo, newFakeTimer())

	err := uc.DeleteOffering(reportAdminCtx(), testOfferingID)

	if !errors.Is(err, errAuditWrite) {
		t.Fatalf("expected the audit write error to propagate, got %v", err)
	}
	if len(auditRepo.saved) != 0 {
		t.Errorf("audit logs stored = %d, want 0", len(auditRepo.saved))
	}
}
