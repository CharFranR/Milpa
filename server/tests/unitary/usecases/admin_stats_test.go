package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

type fakeAdminStatsRepo struct {
	counts domain.AdminStats
	err    error
	calls  int
}

func (r *fakeAdminStatsRepo) Counts(ctx context.Context) (domain.AdminStats, error) {
	r.calls++
	return r.counts, r.err
}

func TestAdminStatsGetRefusesANonAdmin(t *testing.T) {
	t.Parallel()

	repo := &fakeAdminStatsRepo{counts: domain.AdminStats{TotalUsers: 7}}
	uc := usecases.NewAdminStatsUseCase(repo)

	_, err := uc.Get(principalCtx())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Get() error = %v, want ErrForbidden", err)
	}
	if repo.calls != 0 {
		t.Errorf("a refused caller read the counts %d times, want 0", repo.calls)
	}
}

func TestAdminStatsGetRequiresAuthentication(t *testing.T) {
	t.Parallel()

	repo := &fakeAdminStatsRepo{counts: domain.AdminStats{TotalUsers: 7}}
	uc := usecases.NewAdminStatsUseCase(repo)

	_, err := uc.Get(context.Background())
	if !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("Get() error = %v, want ErrUnauthenticated", err)
	}
	if repo.calls != 0 {
		t.Errorf("an unauthenticated caller read the counts %d times, want 0", repo.calls)
	}
}

func TestAdminStatsGetMapsAllFiveCounters(t *testing.T) {
	t.Parallel()

	repo := &fakeAdminStatsRepo{counts: domain.AdminStats{
		TotalUsers:       42,
		SuspendedUsers:   3,
		ActiveOfferings:  17,
		OpenLiquidations: 5,
		PendingReports:   2,
	}}
	uc := usecases.NewAdminStatsUseCase(repo)

	result, err := uc.Get(reportAdminCtx())
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if result.TotalUsers != 42 {
		t.Errorf("TotalUsers = %d, want 42", result.TotalUsers)
	}
	if result.SuspendedUsers != 3 {
		t.Errorf("SuspendedUsers = %d, want 3", result.SuspendedUsers)
	}
	if result.ActiveOfferings != 17 {
		t.Errorf("ActiveOfferings = %d, want 17", result.ActiveOfferings)
	}
	if result.OpenLiquidations != 5 {
		t.Errorf("OpenLiquidations = %d, want 5", result.OpenLiquidations)
	}
	if result.PendingReports != 2 {
		t.Errorf("PendingReports = %d, want 2", result.PendingReports)
	}
}

func TestAdminStatsGetSurfacesTheRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("counts unavailable")
	repo := &fakeAdminStatsRepo{err: repoErr}
	uc := usecases.NewAdminStatsUseCase(repo)

	_, err := uc.Get(reportAdminCtx())
	if !errors.Is(err, repoErr) {
		t.Fatalf("Get() error = %v, want the repository error unchanged", err)
	}
}
