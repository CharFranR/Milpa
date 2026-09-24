package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var testReportIDs = []uuid.UUID{
	uuid.MustParse("88888888-8888-8888-8888-888888888801"),
	uuid.MustParse("88888888-8888-8888-8888-888888888802"),
	uuid.MustParse("88888888-8888-8888-8888-888888888803"),
	uuid.MustParse("88888888-8888-8888-8888-888888888804"),
	uuid.MustParse("88888888-8888-8888-8888-888888888805"),
}

var testReportNotFoundID uuid.UUID = uuid.MustParse("88888888-8888-8888-8888-888888888899")

var testReportReporterID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111113")
var testReportAdminID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111114")
var testReportSecondReporterID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111115")

var testReportTargetIDs = []uuid.UUID{
	uuid.MustParse("99999999-9999-9999-9999-999999999901"),
	uuid.MustParse("99999999-9999-9999-9999-999999999902"),
	uuid.MustParse("99999999-9999-9999-9999-999999999903"),
	uuid.MustParse("99999999-9999-9999-9999-999999999904"),
	uuid.MustParse("99999999-9999-9999-9999-999999999905"),
}

var testReportTargetUserID uuid.UUID = uuid.MustParse("99999999-9999-9999-9999-999999999910")

func newReportFixture(id, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID, status domain.ReportStatus, createdAt time.Time) *domain.Report {
	return &domain.Report{
		ID:         id,
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     "Fraudulent listing pretending to be a government grant",
		Status:     status,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}
}

func setupReportTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testReportReporterID,
			FirstName:    "Report",
			LastName:     "Reporter",
			Role:         domain.RoleMIPYME,
			Email:        "report-reporter@example.com",
			PhoneNumber:  "1111-1111",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testReportAdminID,
			FirstName:    "Moderation",
			LastName:     "Admin",
			Role:         domain.RoleAdmin,
			Email:        "report-admin@example.com",
			PhoneNumber:  "2222-2222",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testReportSecondReporterID,
			FirstName:    "Second",
			LastName:     "Reporter",
			Role:         domain.RoleMIPYME,
			Email:        "report-reporter-2@example.com",
			PhoneNumber:  "3333-3333",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}

	for _, u := range users {
		if _, err := userRepo.Save(context.Background(), u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}
}

func TestReportSaveAndFindByID(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)

	saved := newReportFixture(testReportIDs[0], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime)
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testReportIDs[0],
			ExpectedErr: nil,
		},
		{
			Name:        "Report Not Found",
			ID:          testReportNotFoundID,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			report, err := db.FindByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByID() unexpected error: %v", err)
				return
			}

			if report == nil {
				t.Fatal("FindByID() returned nil report")
			}
			if report.ID != saved.ID {
				t.Errorf("FindByID() ID = %v, want %v", report.ID, saved.ID)
			}
			if report.ReporterID != saved.ReporterID {
				t.Errorf("FindByID() ReporterID = %v, want %v", report.ReporterID, saved.ReporterID)
			}
			if report.TargetType != saved.TargetType {
				t.Errorf("FindByID() TargetType = %v, want %v", report.TargetType, saved.TargetType)
			}
			if report.TargetID != saved.TargetID {
				t.Errorf("FindByID() TargetID = %v, want %v", report.TargetID, saved.TargetID)
			}
			if report.Reason != saved.Reason {
				t.Errorf("FindByID() Reason = %v, want %v", report.Reason, saved.Reason)
			}
			if report.Status != saved.Status {
				t.Errorf("FindByID() Status = %v, want %v", report.Status, saved.Status)
			}
			if report.ResolvedBy != nil {
				t.Errorf("FindByID() ResolvedBy = %v, want nil for a pending report", report.ResolvedBy)
			}
			if report.ResolvedAt != nil {
				t.Errorf("FindByID() ResolvedAt = %v, want nil for a pending report", report.ResolvedAt)
			}
			if !report.CreatedAt.Equal(saved.CreatedAt) {
				t.Errorf("FindByID() CreatedAt = %v, want %v", report.CreatedAt, saved.CreatedAt)
			}
			if !report.UpdatedAt.Equal(saved.UpdatedAt) {
				t.Errorf("FindByID() UpdatedAt = %v, want %v", report.UpdatedAt, saved.UpdatedAt)
			}
		})
	}
}

func TestReportFindAllPagination(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)

	saved := make([]*domain.Report, 0, len(testReportIDs))
	for i, id := range testReportIDs {
		report := newReportFixture(id, testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[i], domain.ReportPending, fixedTime.Add(time.Duration(i)*time.Minute))
		if err := db.Save(context.Background(), report); err != nil {
			t.Fatalf("Save() report %d: %v", i, err)
		}
		saved = append(saved, report)
	}

	tests := []struct {
		Name          string
		Page          int
		PageSize      int
		ExpectedLen   int
		ExpectedTotal int
		ExpectedIDs   []uuid.UUID
	}{
		{
			Name:          "First page holds the newest reports",
			Page:          1,
			PageSize:      2,
			ExpectedLen:   2,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[4].ID, saved[3].ID},
		},
		{
			Name:          "Second page holds the middle reports",
			Page:          2,
			PageSize:      2,
			ExpectedLen:   2,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[2].ID, saved[1].ID},
		},
		{
			Name:          "Last page holds the oldest report only",
			Page:          3,
			PageSize:      2,
			ExpectedLen:   1,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[0].ID},
		},
		{
			Name:          "Page beyond the last one is empty but keeps the total",
			Page:          4,
			PageSize:      2,
			ExpectedLen:   0,
			ExpectedTotal: 5,
			ExpectedIDs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reports, total, err := db.FindAll(context.Background(), "", "", tt.Page, tt.PageSize)
			if err != nil {
				t.Errorf("FindAll() unexpected error: %v", err)
				return
			}

			if total != tt.ExpectedTotal {
				t.Errorf("FindAll() total = %d, want %d", total, tt.ExpectedTotal)
			}
			if len(reports) != tt.ExpectedLen {
				t.Fatalf("FindAll() got %d reports, want %d", len(reports), tt.ExpectedLen)
			}

			for i, wantID := range tt.ExpectedIDs {
				if reports[i].ID != wantID {
					t.Errorf("FindAll()[%d].ID = %v, want %v", i, reports[i].ID, wantID)
				}
			}
		})
	}
}

func TestReportFindAllFilters(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)

	fixtures := []*domain.Report{
		newReportFixture(testReportIDs[0], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime),
		newReportFixture(testReportIDs[1], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[1], domain.ReportPending, fixedTime.Add(time.Minute)),
		newReportFixture(testReportIDs[2], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportApproved, fixedTime.Add(2*time.Minute)),
		newReportFixture(testReportIDs[3], testReportReporterID, domain.ReportTargetUser, testReportTargetUserID, domain.ReportPending, fixedTime.Add(3*time.Minute)),
		newReportFixture(testReportIDs[4], testReportReporterID, domain.ReportTargetUser, testReportTargetUserID, domain.ReportRejected, fixedTime.Add(4*time.Minute)),
	}

	for _, report := range fixtures {
		if err := db.Save(context.Background(), report); err != nil {
			t.Fatalf("Save() report %s: %v", report.ID, err)
		}
	}

	tests := []struct {
		Name          string
		Status        string
		TargetType    string
		ExpectedTotal int
		ExpectErr     bool
	}{
		{
			Name:          "No filters returns every report",
			ExpectedTotal: 5,
		},
		{
			Name:          "Status pending",
			Status:        "pending",
			ExpectedTotal: 3,
		},
		{
			Name:          "Status approved",
			Status:        "approved",
			ExpectedTotal: 1,
		},
		{
			Name:          "Status rejected",
			Status:        "rejected",
			ExpectedTotal: 1,
		},
		{
			Name:          "Target type offering",
			TargetType:    "offering",
			ExpectedTotal: 3,
		},
		{
			Name:          "Target type user",
			TargetType:    "user",
			ExpectedTotal: 2,
		},
		{
			Name:          "Status and target type combined",
			Status:        "pending",
			TargetType:    "offering",
			ExpectedTotal: 2,
		},
		{
			Name:          "Status and target type with no match",
			Status:        "approved",
			TargetType:    "user",
			ExpectedTotal: 0,
		},
		{
			Name:          "Unknown status fails at the database",
			Status:        "bogus",
			ExpectErr:     true,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reports, total, err := db.FindAll(context.Background(), tt.Status, tt.TargetType, 1, 10)

			if tt.ExpectErr {
				if err == nil {
					t.Error("FindAll() error = nil, want a database error for an unknown enum value")
				}
				return
			}

			if err != nil {
				t.Errorf("FindAll() unexpected error: %v", err)
				return
			}
			if total != tt.ExpectedTotal {
				t.Errorf("FindAll() total = %d, want %d", total, tt.ExpectedTotal)
			}
			if len(reports) != tt.ExpectedTotal {
				t.Errorf("FindAll() got %d reports, want %d", len(reports), tt.ExpectedTotal)
			}
		})
	}
}

func TestReportExistsPendingByTarget(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)
	ctx := context.Background()

	exists, err := db.ExistsPendingByTarget(ctx, testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0])
	if err != nil {
		t.Fatalf("ExistsPendingByTarget() unexpected error: %v", err)
	}
	if exists {
		t.Error("ExistsPendingByTarget() = true with no reports, want false")
	}

	pending := newReportFixture(testReportIDs[0], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime)
	if err := db.Save(ctx, pending); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	exists, err = db.ExistsPendingByTarget(ctx, testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0])
	if err != nil {
		t.Fatalf("ExistsPendingByTarget() unexpected error: %v", err)
	}
	if !exists {
		t.Error("ExistsPendingByTarget() = false after a pending report, want true")
	}

	exists, err = db.ExistsPendingByTarget(ctx, testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[1])
	if err != nil {
		t.Fatalf("ExistsPendingByTarget() unexpected error: %v", err)
	}
	if exists {
		t.Error("ExistsPendingByTarget() = true for a different target, want false")
	}

	duplicate := newReportFixture(testReportIDs[1], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime.Add(time.Minute))
	err = db.Save(ctx, duplicate)
	if err == nil {
		t.Fatal("Save() a second pending report for the same target: error = nil, want a unique index violation")
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("Save() duplicate pending report: error = %v (%T), want *pgconn.PgError", err, err)
	}
	if pgErr.Code != "23505" {
		t.Errorf("Save() duplicate pending report: code = %s, want 23505 (unique_violation)", pgErr.Code)
	}
	if pgErr.ConstraintName != "idx_reports_pending_unique" {
		t.Errorf("Save() duplicate pending report: constraint = %s, want idx_reports_pending_unique", pgErr.ConstraintName)
	}

	otherReporter := newReportFixture(testReportIDs[2], testReportSecondReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime.Add(2*time.Minute))
	if err := db.Save(ctx, otherReporter); err != nil {
		t.Fatalf("Save() pending report from another reporter: %v", err)
	}

	pending.Status = domain.ReportApproved
	if err := db.Resolve(ctx, pending); err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	replacement := newReportFixture(testReportIDs[3], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime.Add(3*time.Minute))
	if err := db.Save(ctx, replacement); err != nil {
		t.Fatalf("Save() after the previous report was resolved: %v", err)
	}
}

func TestReportResolve(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)
	ctx := context.Background()

	tests := []struct {
		Name           string
		ReportIndex    int
		Resolve        func(r *domain.Report, adminID uuid.UUID, now time.Time)
		ExpectedStatus domain.ReportStatus
	}{
		{
			Name:           "Pending to approved",
			ReportIndex:    0,
			Resolve:        (*domain.Report).Approve,
			ExpectedStatus: domain.ReportApproved,
		},
		{
			Name:           "Pending to rejected",
			ReportIndex:    1,
			Resolve:        (*domain.Report).Reject,
			ExpectedStatus: domain.ReportRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			report := newReportFixture(testReportIDs[tt.ReportIndex], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[tt.ReportIndex], domain.ReportPending, fixedTime)
			if err := db.Save(ctx, report); err != nil {
				t.Fatalf("Save() error: %v", err)
			}

			stored, err := db.FindByID(ctx, report.ID)
			if err != nil {
				t.Fatalf("FindByID() error: %v", err)
			}

			tt.Resolve(stored, testReportAdminID, fixedTime2)
			if err := db.Resolve(ctx, stored); err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}

			reloaded, err := db.FindByID(ctx, report.ID)
			if err != nil {
				t.Fatalf("FindByID() after Resolve: %v", err)
			}

			if reloaded.Status != tt.ExpectedStatus {
				t.Errorf("Resolve() Status = %v, want %v", reloaded.Status, tt.ExpectedStatus)
			}
			if reloaded.ResolvedBy == nil {
				t.Fatal("Resolve() ResolvedBy = nil, want the admin id")
			}
			if *reloaded.ResolvedBy != testReportAdminID {
				t.Errorf("Resolve() ResolvedBy = %v, want %v", *reloaded.ResolvedBy, testReportAdminID)
			}
			if reloaded.ResolvedAt == nil {
				t.Fatal("Resolve() ResolvedAt = nil, want the resolution timestamp")
			}
			if !reloaded.ResolvedAt.Equal(fixedTime2) {
				t.Errorf("Resolve() ResolvedAt = %v, want %v", reloaded.ResolvedAt, fixedTime2)
			}
			if !reloaded.UpdatedAt.Equal(fixedTime2) {
				t.Errorf("Resolve() UpdatedAt = %v, want %v", reloaded.UpdatedAt, fixedTime2)
			}
		})
	}
}

func TestReportResolveMissingOrAlreadyResolved(t *testing.T) {
	setupReportTestData(t)
	db := repository.NewReportRepository(TestPool)
	ctx := context.Background()

	t.Run("Resolve on a non-existent report", func(t *testing.T) {
		ghost := newReportFixture(testReportNotFoundID, testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime)

		err := db.Resolve(ctx, ghost)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Resolve() on a missing report error = %v, want %v", err, domain.ErrNotFound)
		}

		_, findErr := db.FindByID(ctx, testReportNotFoundID)
		if !errors.Is(findErr, domain.ErrNotFound) {
			t.Errorf("FindByID() after resolving a missing report error = %v, want %v", findErr, domain.ErrNotFound)
		}
	})

	t.Run("Resolve on an already-resolved report", func(t *testing.T) {
		report := newReportFixture(testReportIDs[0], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime)
		if err := db.Save(ctx, report); err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		stored, err := db.FindByID(ctx, report.ID)
		if err != nil {
			t.Fatalf("FindByID() error: %v", err)
		}
		stored.Approve(testReportAdminID, fixedTime2)
		if err := db.Resolve(ctx, stored); err != nil {
			t.Fatalf("Resolve() first call error: %v", err)
		}

		second, err := db.FindByID(ctx, report.ID)
		if err != nil {
			t.Fatalf("FindByID() error: %v", err)
		}
		second.Reject(testReportAdminID, fixedTime2.Add(time.Hour))

		err = db.Resolve(ctx, second)
		if !errors.Is(err, domain.ErrReportAlreadyResolved) {
			t.Errorf("Resolve() on an already-resolved report error = %v, want %v", err, domain.ErrReportAlreadyResolved)
		}

		final, err := db.FindByID(ctx, report.ID)
		if err != nil {
			t.Fatalf("FindByID() after second Resolve: %v", err)
		}
		if final.Status != domain.ReportApproved {
			t.Errorf("Resolve() second call Status = %v, want %v (the terminal state must not be overwritten)", final.Status, domain.ReportApproved)
		}
		if final.ResolvedAt == nil || !final.ResolvedAt.Equal(fixedTime2) {
			t.Errorf("Resolve() second call ResolvedAt = %v, want %v (the first resolution must be preserved)", final.ResolvedAt, fixedTime2)
		}
	})
}

func TestReportResolvePendingByTarget(t *testing.T) {
	tests := []struct {
		Name           string
		Resolve        func(r *domain.Report, adminID uuid.UUID, now time.Time)
		ExpectedStatus domain.ReportStatus
	}{
		{
			Name:           "Approve closes every pending report for the target",
			Resolve:        (*domain.Report).Approve,
			ExpectedStatus: domain.ReportApproved,
		},
		{
			Name:           "Reject closes every pending report for the target",
			Resolve:        (*domain.Report).Reject,
			ExpectedStatus: domain.ReportRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			setupReportTestData(t)
			db := repository.NewReportRepository(TestPool)
			ctx := context.Background()

			primary := newReportFixture(testReportIDs[0], testReportReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime)
			sibling := newReportFixture(testReportIDs[1], testReportSecondReporterID, domain.ReportTargetOffering, testReportTargetIDs[0], domain.ReportPending, fixedTime.Add(time.Minute))
			otherTarget := newReportFixture(testReportIDs[2], testReportSecondReporterID, domain.ReportTargetOffering, testReportTargetIDs[1], domain.ReportPending, fixedTime.Add(2*time.Minute))

			for _, report := range []*domain.Report{primary, sibling, otherTarget} {
				if err := db.Save(ctx, report); err != nil {
					t.Fatalf("Save() %s: %v", report.ID, err)
				}
			}

			stored, err := db.FindByID(ctx, primary.ID)
			if err != nil {
				t.Fatalf("FindByID() error: %v", err)
			}

			tt.Resolve(stored, testReportAdminID, fixedTime2)
			if err := db.Resolve(ctx, stored); err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}
			if err := db.ResolvePendingByTarget(ctx, stored); err != nil {
				t.Fatalf("ResolvePendingByTarget() error: %v", err)
			}

			closed, err := db.FindByID(ctx, sibling.ID)
			if err != nil {
				t.Fatalf("FindByID() sibling: %v", err)
			}
			if closed.Status != tt.ExpectedStatus {
				t.Errorf("sibling Status = %v, want %v", closed.Status, tt.ExpectedStatus)
			}
			if closed.ResolvedBy == nil {
				t.Fatal("sibling ResolvedBy = nil, want the admin id")
			}
			if *closed.ResolvedBy != testReportAdminID {
				t.Errorf("sibling ResolvedBy = %v, want %v", *closed.ResolvedBy, testReportAdminID)
			}
			if closed.ResolvedAt == nil {
				t.Fatal("sibling ResolvedAt = nil, want the resolution timestamp")
			}
			if !closed.ResolvedAt.Equal(fixedTime2) {
				t.Errorf("sibling ResolvedAt = %v, want %v", closed.ResolvedAt, fixedTime2)
			}
			if !closed.UpdatedAt.Equal(fixedTime2) {
				t.Errorf("sibling UpdatedAt = %v, want %v", closed.UpdatedAt, fixedTime2)
			}

			untouched, err := db.FindByID(ctx, otherTarget.ID)
			if err != nil {
				t.Fatalf("FindByID() report on a different target: %v", err)
			}
			if untouched.Status != domain.ReportPending {
				t.Errorf("report on a different target Status = %v, want %v", untouched.Status, domain.ReportPending)
			}
			if untouched.ResolvedBy != nil {
				t.Errorf("report on a different target ResolvedBy = %v, want nil", untouched.ResolvedBy)
			}
			if untouched.ResolvedAt != nil {
				t.Errorf("report on a different target ResolvedAt = %v, want nil", untouched.ResolvedAt)
			}
			if !untouched.UpdatedAt.Equal(otherTarget.UpdatedAt) {
				t.Errorf("report on a different target UpdatedAt = %v, want %v", untouched.UpdatedAt, otherTarget.UpdatedAt)
			}
		})
	}
}
