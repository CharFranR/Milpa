package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

var testResolveReportID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222221")
var testResolveReporterID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
var testResolveAdminID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222223")
var testResolveTargetID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222224")

func newResolvedReportFixture(status domain.ReportStatus) *domain.Report {
	return &domain.Report{
		ID:         testResolveReportID,
		ReporterID: testResolveReporterID,
		TargetType: domain.ReportTargetOffering,
		TargetID:   testResolveTargetID,
		Reason:     "Fraudulent listing pretending to be a government grant",
		Status:     status,
		ResolvedBy: &testResolveAdminID,
		ResolvedAt: &fixedTime2,
		CreatedAt:  fixedTime,
		UpdatedAt:  fixedTime2,
	}
}

func TestReportResolveCommitsPrimaryAndSiblingCascadeTogether(t *testing.T) {
	report := newResolvedReportFixture(domain.ReportApproved)

	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("fail to create mock: %v", err)
	}
	defer mockPool.Close()

	repo := repository.NewReportRepository(mockPool)

	mockPool.ExpectBegin()
	mockPool.ExpectExec("UPDATE reports").
		WithArgs(report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectExec("UPDATE reports").
		WithArgs(report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.TargetType, report.TargetID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectCommit()

	if err := repo.Resolve(context.Background(), report); err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestReportResolveRollsBackWhenSiblingCascadeFails(t *testing.T) {
	report := newResolvedReportFixture(domain.ReportApproved)
	cascadeErr := errors.New("sibling cascade failed")

	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("fail to create mock: %v", err)
	}
	defer mockPool.Close()

	repo := repository.NewReportRepository(mockPool)

	mockPool.ExpectBegin()
	mockPool.ExpectExec("UPDATE reports").
		WithArgs(report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectExec("UPDATE reports").
		WithArgs(report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.TargetType, report.TargetID).
		WillReturnError(cascadeErr)
	mockPool.ExpectRollback()

	err = repo.Resolve(context.Background(), report)
	if !errors.Is(err, cascadeErr) {
		t.Errorf("Resolve() error = %v, want %v", err, cascadeErr)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations (the transaction must roll back and never commit): %v", err)
	}
}

func TestReportResolveMissRollsBackBeforeReadingStatus(t *testing.T) {
	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantErr error
	}{
		{
			name:    "report does not exist",
			rows:    pgxmock.NewRows([]string{"status"}),
			wantErr: domain.ErrNotFound,
		},
		{
			name:    "report already holds a terminal state",
			rows:    pgxmock.NewRows([]string{"status"}).AddRow(domain.ReportApproved),
			wantErr: domain.ErrReportAlreadyResolved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := newResolvedReportFixture(domain.ReportApproved)

			mockPool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("fail to create mock: %v", err)
			}
			defer mockPool.Close()

			repo := repository.NewReportRepository(mockPool)

			mockPool.ExpectBegin()
			mockPool.ExpectExec("UPDATE reports").
				WithArgs(report.Status, report.ResolvedBy, report.ResolvedAt, report.UpdatedAt, report.ID).
				WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			mockPool.ExpectRollback()
			mockPool.ExpectQuery("SELECT status FROM reports").
				WithArgs(report.ID).
				WillReturnRows(tt.rows)

			err = repo.Resolve(context.Background(), report)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Resolve() error = %v, want %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations (the transaction must roll back before the status read): %v", err)
			}
		})
	}
}
