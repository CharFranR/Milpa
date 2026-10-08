package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

var adminStatsAdminID = uuid.MustParse("46464646-4646-4646-4646-464646464646")

type adminStatsRepo struct {
	counts domain.AdminStats
	err    error
}

func (r *adminStatsRepo) Counts(ctx context.Context) (domain.AdminStats, error) {
	return r.counts, r.err
}

func newAdminStatsRouter(counts domain.AdminStats) http.Handler {
	uc := usecases.NewAdminStatsUseCase(&adminStatsRepo{counts: counts})

	return NewRouter(
		nil, nil, nil, nil, nil, nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil,
		handler.NewAdminStatsHandler(uc),
	)
}

func TestAdminStatsRouteRequiresAuthentication(t *testing.T) {
	t.Parallel()

	router := newAdminStatsRouter(domain.AdminStats{TotalUsers: 9})

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
	}
}

func TestAdminStatsRouteRefusesANonAdmin(t *testing.T) {
	t.Parallel()

	router := newAdminStatsRouter(domain.AdminStats{TotalUsers: 9})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(adminStatsAdminID, domain.RoleAgricultor))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}
}

func TestAdminStatsRouteRefusesAnAuditorInTheUseCase(t *testing.T) {
	t.Parallel()

	router := newAdminStatsRouter(domain.AdminStats{TotalUsers: 9})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(adminStatsAdminID, domain.RoleAuditor))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 from the use case; body = %s", rr.Code, rr.Body.String())
	}
}

func TestAdminStatsRouteReturnsAllFiveCounters(t *testing.T) {
	t.Parallel()

	router := newAdminStatsRouter(domain.AdminStats{
		TotalUsers:       42,
		SuspendedUsers:   3,
		ActiveOfferings:  17,
		OpenLiquidations: 5,
		PendingReports:   2,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(adminStatsAdminID, domain.RoleAdmin))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v; body = %s", err, rr.Body.String())
	}
	if len(body.Data) != 5 {
		t.Fatalf("body has %d keys, want 5; body = %s", len(body.Data), rr.Body.String())
	}

	want := map[string]float64{
		"total_users":       42,
		"suspended_users":   3,
		"active_offerings":  17,
		"open_liquidations": 5,
		"pending_reports":   2,
	}
	for key, value := range want {
		got, ok := body.Data[key]
		if !ok {
			t.Errorf("body is missing the %q key: %s", key, rr.Body.String())
			continue
		}
		if got != value {
			t.Errorf("%s = %v, want %v", key, got, value)
		}
	}
}
