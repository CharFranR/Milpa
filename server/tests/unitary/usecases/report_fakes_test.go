package usecases_test

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

var (
	testReportID = uuid.MustParse("88888888-8888-8888-8888-888888888888")
	testAdminID  = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	testTargetID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
)

func adminCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: testAdminID, Role: domain.RoleAdmin})
}

func mustReport() *domain.Report {
	report, err := domain.NewReport(testUserID, domain.ReportTargetOffering, testTargetID, "Este producto parece fraudulento, no existe", fixedTime)
	if err != nil {
		panic(err)
	}
	report.ID = testReportID
	return report
}

type fakeReportRepo struct {
	save          func(ctx context.Context, report *domain.Report) error
	findByID      func(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	findAll       func(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error)
	resolve       func(ctx context.Context, report *domain.Report) error
	existsPending func(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error)
	saved         []*domain.Report
	resolved      []*domain.Report
}

func newFakeReportRepo() *fakeReportRepo {
	f := &fakeReportRepo{}
	f.save = func(ctx context.Context, report *domain.Report) error {
		f.saved = append(f.saved, report)
		return nil
	}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
		report := mustReport()
		report.ID = id
		return report, nil
	}
	f.findAll = func(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
		return []domain.Report{*mustReport()}, 1, nil
	}
	f.resolve = func(ctx context.Context, report *domain.Report) error {
		f.resolved = append(f.resolved, report)
		return nil
	}
	f.existsPending = func(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error) {
		return false, nil
	}
	return f
}

func (f *fakeReportRepo) Save(ctx context.Context, report *domain.Report) error {
	return f.save(ctx, report)
}

func (f *fakeReportRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	return f.findByID(ctx, id)
}

func (f *fakeReportRepo) FindAll(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
	return f.findAll(ctx, status, targetType, page, pageSize)
}

func (f *fakeReportRepo) Resolve(ctx context.Context, report *domain.Report) error {
	return f.resolve(ctx, report)
}

func (f *fakeReportRepo) ExistsPendingByTarget(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error) {
	return f.existsPending(ctx, reporterID, targetType, targetID)
}

type fakeAuditLogRepo struct {
	save    func(ctx context.Context, log *domain.AuditLog) error
	findAll func(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error)
	saved   []*domain.AuditLog
}

func newFakeAuditLogRepo() *fakeAuditLogRepo {
	f := &fakeAuditLogRepo{}
	f.save = func(ctx context.Context, log *domain.AuditLog) error {
		f.saved = append(f.saved, log)
		return nil
	}
	f.findAll = func(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error) {
		return []domain.AuditLog{{
			ID:         uuid.New(),
			ActorID:    testAdminID,
			Action:     domain.AuditActionReportCreated,
			TargetType: "report",
			TargetID:   testReportID,
			Metadata:   json.RawMessage(`{}`),
			CreatedAt:  fixedTime,
		}}, 1, nil
	}
	return f
}

func (f *fakeAuditLogRepo) Save(ctx context.Context, log *domain.AuditLog) error {
	return f.save(ctx, log)
}

func (f *fakeAuditLogRepo) FindAll(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error) {
	return f.findAll(ctx, action, actorID, targetType, page, pageSize)
}

func (f *fakeAuditLogRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	return &domain.AuditLog{}, nil
}
