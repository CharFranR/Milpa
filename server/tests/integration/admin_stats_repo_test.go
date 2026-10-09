package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

func adminStatsInsertUser(t *testing.T, id uuid.UUID, suspended bool) {
	t.Helper()

	var suspendedAt *time.Time
	if suspended {
		now := time.Now().UTC()
		suspendedAt = &now
	}

	if _, err := TestPool.Exec(context.Background(),
		`INSERT INTO users (id, first_name, last_name, role, email, phone_number, password_hash, created_at, updated_at, suspended_at)
		 VALUES ($1, 'Stats', 'User', 1, $2, '555-0000', 'hash', NOW(), NOW(), $3)`,
		id, id.String()+"@milpa.com.ni", suspendedAt,
	); err != nil {
		t.Fatalf("insert user %s: %v", id, err)
	}
}

func adminStatsInsertOffering(t *testing.T, id, userID uuid.UUID, isActive bool, expiresAt *time.Time) {
	t.Helper()

	if _, err := TestPool.Exec(context.Background(),
		`INSERT INTO offerings (id, user_id, type, name, price, is_active, expires_at, created_at, updated_at)
		 VALUES ($1, $2, 1, $3, 10, $4, $5, NOW(), NOW())`,
		id, userID, "Oferta "+id.String()[9:13], isActive, expiresAt,
	); err != nil {
		t.Fatalf("insert offering %s: %v", id, err)
	}
}

func adminStatsInsertLiquidation(t *testing.T, id, supplierID uuid.UUID, status string) {
	t.Helper()

	if _, err := TestPool.Exec(context.Background(),
		`INSERT INTO liquidations (id, supplier_id, product_name, quantity, unit_of_measure, total_price, unit_price, location_id, status, created_at, updated_at)
		 VALUES ($1, $2, 'Cafe', 10, 'kg', 100, 10, $3, $4, NOW(), NOW())`,
		id, supplierID, uuid.New(), status,
	); err != nil {
		t.Fatalf("insert liquidation %s: %v", id, err)
	}
}

func adminStatsInsertReport(t *testing.T, id, reporterID, targetID uuid.UUID, status string) {
	t.Helper()

	if _, err := TestPool.Exec(context.Background(),
		`INSERT INTO reports (id, reporter_id, target_type, target_id, reason, status, created_at, updated_at)
		 VALUES ($1, $2, 'offering', $3, 'Este producto parece fraudulento', $4, NOW(), NOW())`,
		id, reporterID, targetID, status,
	); err != nil {
		t.Fatalf("insert report %s: %v", id, err)
	}
}

func TestAdminStatsCountsAgainstPostgres(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewAdminStatsRepository(TestPool)
	ctx := context.Background()

	activeUser := uuid.MustParse("aaaaaaaa-1111-4111-8111-111111111111")
	suspendedUser := uuid.MustParse("aaaaaaaa-2222-4222-8222-222222222222")
	adminStatsInsertUser(t, activeUser, false)
	adminStatsInsertUser(t, suspendedUser, true)

	future := time.Now().UTC().Add(24 * time.Hour)
	past := time.Now().UTC().Add(-24 * time.Hour)

	adminStatsInsertOffering(t, uuid.MustParse("bbbbbbbb-1111-4111-8111-111111111111"), activeUser, true, &future)
	adminStatsInsertOffering(t, uuid.MustParse("bbbbbbbb-2222-4222-8222-222222222222"), activeUser, false, nil)
	adminStatsInsertOffering(t, uuid.MustParse("bbbbbbbb-3333-4333-8333-333333333333"), activeUser, true, &past)

	adminStatsInsertLiquidation(t, uuid.MustParse("cccccccc-1111-4111-8111-111111111111"), activeUser, "open")
	adminStatsInsertLiquidation(t, uuid.MustParse("cccccccc-2222-4222-8222-222222222222"), activeUser, "closed")

	adminStatsInsertReport(t, uuid.MustParse("dddddddd-1111-4111-8111-111111111111"), activeUser, uuid.MustParse("eeeeeeee-1111-4111-8111-111111111111"), "pending")
	adminStatsInsertReport(t, uuid.MustParse("dddddddd-2222-4222-8222-222222222222"), activeUser, uuid.MustParse("eeeeeeee-2222-4222-8222-222222222222"), "approved")

	stats, err := repo.Counts(ctx)
	if err != nil {
		t.Fatalf("Counts() error: %v", err)
	}

	if stats.TotalUsers != 2 {
		t.Errorf("TotalUsers = %d, want 2", stats.TotalUsers)
	}
	if stats.SuspendedUsers != 1 {
		t.Errorf("SuspendedUsers = %d, want 1", stats.SuspendedUsers)
	}
	if stats.ActiveOfferings != 1 {
		t.Errorf("ActiveOfferings = %d, want only the active unexpired offering", stats.ActiveOfferings)
	}
	if stats.OpenLiquidations != 1 {
		t.Errorf("OpenLiquidations = %d, want 1", stats.OpenLiquidations)
	}
	if stats.PendingReports != 1 {
		t.Errorf("PendingReports = %d, want 1", stats.PendingReports)
	}
}

func TestAdminStatsCountsAreZeroOnAnEmptyDatabase(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewAdminStatsRepository(TestPool)

	stats, err := repo.Counts(context.Background())
	if err != nil {
		t.Fatalf("Counts() error: %v", err)
	}
	if stats != (domain.AdminStats{}) {
		t.Errorf("Counts() = %+v, want every counter at zero", stats)
	}
}

func TestAdminStatsActiveOfferingsCountsANonExpiringOffering(t *testing.T) {
	cleanupTables(t)

	repo := repository.NewAdminStatsRepository(TestPool)

	user := uuid.MustParse("aaaaaaaa-3333-4333-8333-333333333333")
	adminStatsInsertUser(t, user, false)
	adminStatsInsertOffering(t, uuid.MustParse("bbbbbbbb-4444-4444-8444-444444444444"), user, true, nil)

	stats, err := repo.Counts(context.Background())
	if err != nil {
		t.Fatalf("Counts() error: %v", err)
	}
	if stats.ActiveOfferings != 1 {
		t.Errorf("ActiveOfferings = %d, want the non-expiring active offering", stats.ActiveOfferings)
	}
}
