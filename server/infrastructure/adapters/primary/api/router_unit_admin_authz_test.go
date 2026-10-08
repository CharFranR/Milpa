package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

var unitAdminID = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

type unitAdminRepo struct {
	units   []domain.UnitOfMeasure
	saved   []*domain.UnitOfMeasure
	saveErr error
}

func (r *unitAdminRepo) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	active := make([]domain.UnitOfMeasure, 0, len(r.units))
	for i := range r.units {
		if r.units[i].IsActive {
			active = append(active, r.units[i])
		}
	}
	return active, nil
}

func (r *unitAdminRepo) FindAll(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return r.units, nil
}

func (r *unitAdminRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	for i := range r.units {
		if r.units[i].ID == id {
			copied := r.units[i]
			return &copied, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *unitAdminRepo) Save(ctx context.Context, unit *domain.UnitOfMeasure) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	copied := *unit
	r.saved = append(r.saved, &copied)
	for i := range r.units {
		if r.units[i].ID == unit.ID {
			r.units[i] = copied
			return nil
		}
	}
	r.units = append(r.units, copied)
	return nil
}

func newUnitAdminRouter(t *testing.T) (http.Handler, *unitAdminRepo) {
	t.Helper()

	repo := &unitAdminRepo{units: []domain.UnitOfMeasure{
		{ID: unitAdminID, Code: "kg", Name: "Kilogramo", IsActive: true},
	}}

	return newUnitAdminRouterWithRepo(repo), repo
}

func newUnitAdminRouterWithRepo(repo *unitAdminRepo) http.Handler {
	uc := usecases.NewUnitOfMeasureUseCase(repo)

	return NewRouter(
		nil, nil,
		nil,
		nil,
		nil,
		nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		handler.NewUnitOfMeasureHandler(uc),
		nil,
	)
}

func TestUnitOfMeasureAdminRoutesRefuseANonAdmin(t *testing.T) {
	t.Parallel()

	verbs := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list all", http.MethodGet, "/api/v1/admin/units-of-measure", ""},
		{"create", http.MethodPost, "/api/v1/admin/units-of-measure", `{"code":"tn","name":"Tonelada"}`},
		{"update", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String(), `{"name":"Kilo"}`},
		{"status", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String() + "/status", `{"is_active":false}`},
	}

	for _, verb := range verbs {
		t.Run(verb.name, func(t *testing.T) {
			t.Parallel()

			router, repo := newUnitAdminRouter(t)

			req := httptest.NewRequest(verb.method, verb.path, strings.NewReader(verb.body))
			req.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAgricultor))
			if verb.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
			}
			if len(repo.saved) != 0 {
				t.Errorf("a refused caller wrote %d units, want 0", len(repo.saved))
			}
		})
	}
}

func TestUnitOfMeasureAdminRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	verbs := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list all", http.MethodGet, "/api/v1/admin/units-of-measure", ""},
		{"create", http.MethodPost, "/api/v1/admin/units-of-measure", `{"code":"tn","name":"Tonelada"}`},
		{"update", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String(), `{"name":"Kilo"}`},
		{"status", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String() + "/status", `{"is_active":false}`},
	}

	for _, verb := range verbs {
		t.Run(verb.name, func(t *testing.T) {
			t.Parallel()

			router, repo := newUnitAdminRouter(t)

			req := httptest.NewRequest(verb.method, verb.path, strings.NewReader(verb.body))
			if verb.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
			}
			if len(repo.saved) != 0 {
				t.Errorf("an unauthenticated caller wrote %d units, want 0", len(repo.saved))
			}
		})
	}
}

func TestUnitOfMeasureAdminListRefusesAnAuditorInTheUseCase(t *testing.T) {
	t.Parallel()

	router, repo := newUnitAdminRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/units-of-measure", nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAuditor))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 from the use case; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 0 {
		t.Errorf("a refused auditor wrote %d units, want 0", len(repo.saved))
	}
}

func TestUnitOfMeasureCreateRefusesADuplicateCode(t *testing.T) {
	t.Parallel()

	repo := &unitAdminRepo{saveErr: domain.ErrDuplicate}
	router := newUnitAdminRouterWithRepo(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/units-of-measure", strings.NewReader(`{"code":"tn","name":"Tonelada"}`))
	req.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 0 {
		t.Errorf("a refused duplicate was written %d times, want 0", len(repo.saved))
	}
}

func TestUnitOfMeasureCreateRejectsOverLongValuesAtTheDomain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		code string
		unit string
	}{
		{"over-long code", strings.Repeat("a", 21), "Tonelada"},
		{"over-long name", "tn", strings.Repeat("b", 61)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			router, repo := newUnitAdminRouter(t)

			body := `{"code":"` + tc.code + `","name":"` + tc.unit + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/units-of-measure", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAdmin))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
			}
			if len(repo.saved) != 0 {
				t.Errorf("an invalid unit was written %d times, want 0", len(repo.saved))
			}
		})
	}
}

func TestUnitOfMeasureAdminLifecycleReachesTheDomain(t *testing.T) {
	t.Parallel()

	router, repo := newUnitAdminRouter(t)

	created := httptest.NewRequest(http.MethodPost, "/api/v1/admin/units-of-measure", strings.NewReader(`{"code":"  TN  ","name":"  Tonelada  "}`))
	created.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAdmin))
	created.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, created)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 1 {
		t.Fatalf("create wrote %d units, want 1", len(repo.saved))
	}
	if repo.saved[0].Code != "tn" || repo.saved[0].Name != "Tonelada" || !repo.saved[0].IsActive {
		t.Errorf("created = %+v, want the normalized tn/Tonelada active unit", repo.saved[0])
	}

	status := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/units-of-measure/"+unitAdminID.String()+"/status", strings.NewReader(`{"is_active":false}`))
	status.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAdmin))
	status.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, status)

	if rr.Code != http.StatusOK {
		t.Fatalf("status toggle = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 2 || repo.saved[1].IsActive {
		t.Fatalf("after the toggle the writes are %+v, want the unit deactivated", repo.saved)
	}

	listAll := httptest.NewRequest(http.MethodGet, "/api/v1/admin/units-of-measure", nil)
	listAll.Header.Set("Authorization", "Bearer "+piiToken(unitAdminID, domain.RoleAdmin))

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, listAll)

	if rr.Code != http.StatusOK {
		t.Fatalf("list all status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Kilogramo") {
		t.Errorf("the admin list dropped the deactivated unit: %s", rr.Body.String())
	}
}
