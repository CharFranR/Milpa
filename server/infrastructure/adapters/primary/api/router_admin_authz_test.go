package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	primary "milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

var (
	adminAuthzAdminID    = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	adminAuthzTargetID   = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	adminAuthzOfferingID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	adminAuthzReportID   = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

type adminAuthzClock struct{}

func (adminAuthzClock) Now() time.Time {
	return time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
}

type adminAuthzUserRepo struct {
	users   []domain.User
	updated []*domain.User
}

func (r *adminAuthzUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for i := range r.users {
		if r.users[i].ID == id {
			copied := r.users[i]
			return &copied, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *adminAuthzUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (r *adminAuthzUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (r *adminAuthzUserRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (r *adminAuthzUserRepo) Save(ctx context.Context, user *domain.User) (string, error) {
	return "", nil
}

func (r *adminAuthzUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.updated = append(r.updated, user)
	return nil
}

func (r *adminAuthzUserRepo) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	return r.users, len(r.users), nil
}

type adminAuthzOfferingRepo struct {
	deleted []uuid.UUID
}

func (r *adminAuthzOfferingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
	return &domain.Offering{ID: id}, nil
}

func (r *adminAuthzOfferingRepo) FindByUserID(ctx context.Context, companyID uuid.UUID, includeHidden bool) ([]domain.Offering, error) {
	return nil, nil
}

func (r *adminAuthzOfferingRepo) FindAll(ctx context.Context) ([]domain.Offering, error) {
	return nil, nil
}

func (r *adminAuthzOfferingRepo) Save(ctx context.Context, offering *domain.Offering) error {
	return nil
}

func (r *adminAuthzOfferingRepo) Update(ctx context.Context, offering *domain.Offering) error {
	return nil
}

func (r *adminAuthzOfferingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *adminAuthzOfferingRepo) DeactivateExpired(ctx context.Context, now time.Time) ([]domain.Offering, error) {
	return nil, nil
}

type adminAuthzAuditRepo struct {
	saved []*domain.AuditLog
}

func (r *adminAuthzAuditRepo) Save(ctx context.Context, log *domain.AuditLog) error {
	r.saved = append(r.saved, log)
	return nil
}

func (r *adminAuthzAuditRepo) FindAll(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error) {
	return nil, 0, nil
}

type adminAuthzReportRepo struct {
	report   *domain.Report
	resolved []*domain.Report
}

func (r *adminAuthzReportRepo) Save(ctx context.Context, report *domain.Report) error {
	return nil
}

func (r *adminAuthzReportRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	if r.report == nil {
		return nil, domain.ErrNotFound
	}
	copied := *r.report
	copied.ID = id
	return &copied, nil
}

func (r *adminAuthzReportRepo) FindAll(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error) {
	return nil, 0, nil
}

func (r *adminAuthzReportRepo) Resolve(ctx context.Context, report *domain.Report) error {
	r.resolved = append(r.resolved, report)
	return nil
}

func (r *adminAuthzReportRepo) ExistsPendingByTarget(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error) {
	return false, nil
}

func adminAuthzTargetUser() domain.User {
	return domain.User{
		ID:        adminAuthzTargetID,
		FirstName: "Ana",
		LastName:  "Diaz",
		Email:     "ana@milpa.com.ni",
		Role:      domain.RoleAgricultor,
	}
}

func adminAuthzPendingReport() *domain.Report {
	report, err := domain.NewReport(adminAuthzTargetID, domain.ReportTargetOffering, adminAuthzOfferingID, "Este producto parece fraudulento, no existe", adminAuthzClock{}.Now())
	if err != nil {
		panic(err)
	}
	report.ID = adminAuthzReportID
	return report
}

type adminAuthzWorld struct {
	router       http.Handler
	userRepo     *adminAuthzUserRepo
	offeringRepo *adminAuthzOfferingRepo
	reportRepo   *adminAuthzReportRepo
	auditRepo    *adminAuthzAuditRepo
}

func (w *adminAuthzWorld) mutations() int {
	return len(w.userRepo.updated) + len(w.offeringRepo.deleted) + len(w.reportRepo.resolved) + len(w.auditRepo.saved)
}

func newAdminAuthzWorld(t *testing.T) *adminAuthzWorld {
	t.Helper()

	userRepo := &adminAuthzUserRepo{users: []domain.User{adminAuthzTargetUser()}}
	offeringRepo := &adminAuthzOfferingRepo{}
	auditRepo := &adminAuthzAuditRepo{}
	reportRepo := &adminAuthzReportRepo{report: adminAuthzPendingReport()}

	moderation := usecases.NewModerationUseCase(userRepo, offeringRepo, auditRepo, adminAuthzClock{})
	report := usecases.NewReportUseCase(reportRepo, auditRepo, userRepo, offeringRepo, adminAuthzClock{})

	router := NewRouter(
		nil, nil,
		nil,
		nil,
		nil,
		nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil,
		handler.NewReportHandler(report),
		handler.NewModerationHandler(moderation),
		nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil,
		nil,
	)

	return &adminAuthzWorld{
		router:       router,
		userRepo:     userRepo,
		offeringRepo: offeringRepo,
		reportRepo:   reportRepo,
		auditRepo:    auditRepo,
	}
}

type adminAuthzRoute struct {
	name   string
	method string
	path   string
	body   string
}

func adminAuthzMutationRoutes() []adminAuthzRoute {
	return []adminAuthzRoute{
		{"suspend user", http.MethodPatch, "/api/v1/admin/users/" + adminAuthzTargetID.String() + "/suspend", `{"action":"suspend"}`},
		{"set role", http.MethodPatch, "/api/v1/admin/users/" + adminAuthzTargetID.String() + "/role", `{"role":1}`},
		{"delete offering", http.MethodDelete, "/api/v1/admin/offerings/" + adminAuthzOfferingID.String(), ""},
		{"resolve report", http.MethodPatch, "/api/v1/reports/" + adminAuthzReportID.String() + "/action", `{"action":"approve"}`},
	}
}

func adminAuthzRequest(route adminAuthzRoute, role *domain.RoleOptions) *http.Request {
	req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
	if route.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if role != nil {
		req.Header.Set("Authorization", "Bearer "+piiToken(adminAuthzAdminID, *role))
	}
	return req
}

func TestAdminMutationsRequireAuthentication(t *testing.T) {
	t.Parallel()

	for _, route := range adminAuthzMutationRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			world := newAdminAuthzWorld(t)

			rr := httptest.NewRecorder()
			world.router.ServeHTTP(rr, adminAuthzRequest(route, nil))

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
			}
			if world.mutations() != 0 {
				t.Errorf("an unauthenticated caller produced %d mutations, want 0", world.mutations())
			}
		})
	}
}

func TestAdminMutationsRefuseANonAdmin(t *testing.T) {
	t.Parallel()

	farmer := domain.RoleAgricultor
	for _, route := range adminAuthzMutationRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			world := newAdminAuthzWorld(t)

			rr := httptest.NewRecorder()
			world.router.ServeHTTP(rr, adminAuthzRequest(route, &farmer))

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
			}
			if world.mutations() != 0 {
				t.Errorf("a refused caller produced %d mutations, want 0", world.mutations())
			}
		})
	}
}

func TestAdminMutationsRefuseAnAuditor(t *testing.T) {
	t.Parallel()

	auditor := domain.RoleAuditor
	for _, route := range adminAuthzMutationRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			world := newAdminAuthzWorld(t)

			rr := httptest.NewRecorder()
			world.router.ServeHTTP(rr, adminAuthzRequest(route, &auditor))

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
			}
			if world.mutations() != 0 {
				t.Errorf("an auditor produced %d mutations, want 0", world.mutations())
			}
		})
	}
}

func TestAdminSuspendUserReachesTheDomain(t *testing.T) {
	t.Parallel()

	world := newAdminAuthzWorld(t)
	admin := domain.RoleAdmin

	rr := httptest.NewRecorder()
	world.router.ServeHTTP(rr, adminAuthzRequest(adminAuthzMutationRoutes()[0], &admin))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if len(world.userRepo.updated) != 1 {
		t.Fatalf("users updated = %d, want 1", len(world.userRepo.updated))
	}
	if world.userRepo.updated[0].ID != adminAuthzTargetID || !world.userRepo.updated[0].IsSuspended() {
		t.Errorf("persisted = %+v, want the target suspended", world.userRepo.updated[0])
	}
}

func TestAdminSetUserRoleReachesTheDomain(t *testing.T) {
	t.Parallel()

	world := newAdminAuthzWorld(t)
	admin := domain.RoleAdmin

	rr := httptest.NewRecorder()
	world.router.ServeHTTP(rr, adminAuthzRequest(adminAuthzMutationRoutes()[1], &admin))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if len(world.userRepo.updated) != 1 {
		t.Fatalf("users updated = %d, want 1", len(world.userRepo.updated))
	}
	if world.userRepo.updated[0].ID != adminAuthzTargetID || world.userRepo.updated[0].Role != domain.RoleAgricultor {
		t.Errorf("persisted = %+v, want the target promoted to agricultor", world.userRepo.updated[0])
	}
}

func TestAdminDeleteOfferingReachesTheDomain(t *testing.T) {
	t.Parallel()

	world := newAdminAuthzWorld(t)
	admin := domain.RoleAdmin

	rr := httptest.NewRecorder()
	world.router.ServeHTTP(rr, adminAuthzRequest(adminAuthzMutationRoutes()[2], &admin))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", rr.Code, rr.Body.String())
	}
	if len(world.offeringRepo.deleted) != 1 || world.offeringRepo.deleted[0] != adminAuthzOfferingID {
		t.Fatalf("deleted = %v, want the target offering", world.offeringRepo.deleted)
	}
}

func TestAdminResolveReportReachesTheDomain(t *testing.T) {
	t.Parallel()

	world := newAdminAuthzWorld(t)
	admin := domain.RoleAdmin

	rr := httptest.NewRecorder()
	world.router.ServeHTTP(rr, adminAuthzRequest(adminAuthzMutationRoutes()[3], &admin))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if len(world.reportRepo.resolved) != 1 {
		t.Fatalf("reports resolved = %d, want 1", len(world.reportRepo.resolved))
	}
	if len(world.offeringRepo.deleted) != 1 || world.offeringRepo.deleted[0] != adminAuthzOfferingID {
		t.Errorf("deleted = %v, want the reported offering", world.offeringRepo.deleted)
	}
}

func TestAdminReadsReachTheAdminHandlers(t *testing.T) {
	t.Parallel()

	admin := domain.RoleAdmin

	for _, path := range []string{"/api/v1/admin/users", "/api/v1/admin/audit-logs", "/api/v1/reports"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			world := newAdminAuthzWorld(t)

			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+piiToken(adminAuthzAdminID, admin))

			rr := httptest.NewRecorder()
			world.router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestAuditorReadIsRefusedByTheUseCaseNotTheRouteGuard(t *testing.T) {
	t.Parallel()

	world := newAdminAuthzWorld(t)
	auditor := domain.RoleAuditor

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(adminAuthzAdminID, auditor))

	rr := httptest.NewRecorder()
	world.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 from the use case; body = %s", rr.Code, rr.Body.String())
	}
}

type permissiveModerationUC struct{}

func (permissiveModerationUC) SuspendUser(ctx context.Context, id uuid.UUID, req dto.SuspendUserRequest) error {
	return nil
}

func (permissiveModerationUC) SetUserRole(ctx context.Context, id uuid.UUID, req dto.SetUserRoleRequest) error {
	return nil
}

func (permissiveModerationUC) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (permissiveModerationUC) ListAuditLogs(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) (*dto.PaginatedAuditLogsResponse, error) {
	return &dto.PaginatedAuditLogsResponse{}, nil
}

func (permissiveModerationUC) ListUsers(ctx context.Context, page, pageSize int) (*dto.PaginatedUsersResponse, error) {
	return &dto.PaginatedUsersResponse{}, nil
}

type permissiveCategoryUC struct{}

func (permissiveCategoryUC) GetAll(ctx context.Context) ([]*dto.CategoryDTO, error) {
	return nil, nil
}

func (permissiveCategoryUC) Create(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryDTO, error) {
	return &dto.CategoryDTO{}, nil
}

func (permissiveCategoryUC) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCategoryRequest) (*dto.CategoryDTO, error) {
	return &dto.CategoryDTO{}, nil
}

func (permissiveCategoryUC) SetStatus(ctx context.Context, id uuid.UUID, req dto.CategoryStatusRequest) (*dto.CategoryDTO, error) {
	return &dto.CategoryDTO{}, nil
}

type permissiveReportUC struct{}

func (permissiveReportUC) Create(ctx context.Context, req dto.CreateReportRequest) (*dto.ReportResponse, error) {
	return &dto.ReportResponse{}, nil
}

func (permissiveReportUC) List(ctx context.Context, status string, targetType string, page, pageSize int) (*dto.PaginatedReportsResponse, error) {
	return &dto.PaginatedReportsResponse{}, nil
}

func (permissiveReportUC) Resolve(ctx context.Context, id uuid.UUID, req dto.ResolveReportRequest) (*dto.ReportResponse, error) {
	return &dto.ReportResponse{}, nil
}

type permissiveUnitAdminUC struct{}

func (permissiveUnitAdminUC) List(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error) {
	return nil, nil
}

func (permissiveUnitAdminUC) ListAll(ctx context.Context) ([]*dto.UnitOfMeasureDTO, error) {
	return nil, nil
}

func (permissiveUnitAdminUC) Create(ctx context.Context, req dto.CreateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error) {
	return &dto.UnitOfMeasureDTO{}, nil
}

func (permissiveUnitAdminUC) Update(ctx context.Context, id uuid.UUID, req dto.UpdateUnitOfMeasureRequest) (*dto.UnitOfMeasureDTO, error) {
	return &dto.UnitOfMeasureDTO{}, nil
}

func (permissiveUnitAdminUC) SetStatus(ctx context.Context, id uuid.UUID, req dto.UnitOfMeasureStatusRequest) (*dto.UnitOfMeasureDTO, error) {
	return &dto.UnitOfMeasureDTO{}, nil
}

type permissiveStatsUC struct{}

func (permissiveStatsUC) Get(ctx context.Context) (*dto.AdminStatsDTO, error) {
	return &dto.AdminStatsDTO{}, nil
}

var (
	_ primary.ModerationUseCase    = permissiveModerationUC{}
	_ primary.CategoryUseCase      = permissiveCategoryUC{}
	_ primary.ReportUseCase        = permissiveReportUC{}
	_ primary.UnitOfMeasureUseCase = permissiveUnitAdminUC{}
	_ primary.AdminStatsUseCase    = permissiveStatsUC{}
)

func newPermissiveAdminRouter() http.Handler {
	return NewRouter(
		nil, nil, nil, nil,
		handler.NewCategoryHandler(permissiveCategoryUC{}),
		nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil,
		handler.NewReportHandler(permissiveReportUC{}),
		handler.NewModerationHandler(permissiveModerationUC{}),
		nil, nil, nil,
		nil, nil, nil, nil, nil,
		handler.NewUnitOfMeasureHandler(permissiveUnitAdminUC{}),
		handler.NewAdminStatsHandler(permissiveStatsUC{}),
	)
}

func TestAdminMutationRoutesEnforceTheRouteGuardOverPermissiveUseCases(t *testing.T) {
	t.Parallel()

	routes := []adminAuthzRoute{
		{"suspend user", http.MethodPatch, "/api/v1/admin/users/" + adminAuthzTargetID.String() + "/suspend", `{"action":"suspend"}`},
		{"set role", http.MethodPatch, "/api/v1/admin/users/" + adminAuthzTargetID.String() + "/role", `{"role":1}`},
		{"delete offering", http.MethodDelete, "/api/v1/admin/offerings/" + adminAuthzOfferingID.String(), ""},
		{"create category", http.MethodPost, "/api/v1/admin/categories", `{"name":"Insumos"}`},
		{"update category", http.MethodPatch, "/api/v1/admin/categories/" + adminAuthzOfferingID.String(), `{"name":"Insumos"}`},
		{"set category status", http.MethodPatch, "/api/v1/admin/categories/" + adminAuthzOfferingID.String() + "/status", `{"is_active":false}`},
		{"create unit", http.MethodPost, "/api/v1/admin/units-of-measure", `{"code":"tn","name":"Tonelada"}`},
		{"update unit", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String(), `{"name":"Tonelada"}`},
		{"set unit status", http.MethodPatch, "/api/v1/admin/units-of-measure/" + unitAdminID.String() + "/status", `{"is_active":false}`},
		{"resolve report", http.MethodPatch, "/api/v1/reports/" + adminAuthzReportID.String() + "/action", `{"action":"approve"}`},
	}

	farmer := domain.RoleAgricultor
	admin := domain.RoleAdmin

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			router := newPermissiveAdminRouter()

			refused := httptest.NewRecorder()
			router.ServeHTTP(refused, adminAuthzRequest(route, &farmer))
			if refused.Code != http.StatusForbidden {
				t.Fatalf("non-admin status = %d, want 403 from the route guard; body = %s", refused.Code, refused.Body.String())
			}

			allowed := httptest.NewRecorder()
			router.ServeHTTP(allowed, adminAuthzRequest(route, &admin))
			if allowed.Code < 200 || allowed.Code >= 300 {
				t.Fatalf("admin status = %d, want a success from the permissive use case; body = %s", allowed.Code, allowed.Body.String())
			}
		})
	}
}
