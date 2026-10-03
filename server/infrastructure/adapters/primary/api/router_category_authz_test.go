package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

var (
	catalogueCategoryID = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	catalogueAdminID    = uuid.MustParse("99999999-9999-9999-9999-999999999999")
)

type catalogueRepo struct {
	categories []domain.Category
	saved      []*domain.Category
}

func (r *catalogueRepo) FindAll(ctx context.Context) ([]domain.Category, error) {
	return r.categories, nil
}

func (r *catalogueRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	for i := range r.categories {
		if r.categories[i].ID == id {
			copied := r.categories[i]
			return &copied, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *catalogueRepo) Save(ctx context.Context, category *domain.Category) error {
	copied := *category
	r.saved = append(r.saved, &copied)
	for i := range r.categories {
		if r.categories[i].ID == category.ID {
			r.categories[i] = copied
			return nil
		}
	}
	r.categories = append(r.categories, copied)
	return nil
}

func newCatalogueRouter(t *testing.T) (http.Handler, *catalogueRepo) {
	t.Helper()

	repo := &catalogueRepo{}
	uc := usecases.NewCategoryUseCase(repo)

	router := NewRouter(
		nil, nil,
		nil,
		nil,
		handler.NewCategoryHandler(uc),
		nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)

	return router, repo
}

func TestCatalogueWritesRefuseANonAdmin(t *testing.T) {
	t.Parallel()

	verbs := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create", http.MethodPost, "/api/v1/admin/categories", `{"name":"Granos"}`},
		{"update", http.MethodPatch, "/api/v1/admin/categories/" + catalogueCategoryID.String(), `{"name":"Cereales"}`},
		{"status", http.MethodPatch, "/api/v1/admin/categories/" + catalogueCategoryID.String() + "/status", `{"is_active":false}`},
	}

	for _, verb := range verbs {
		t.Run(verb.name, func(t *testing.T) {
			t.Parallel()

			router, repo := newCatalogueRouter(t)

			req := httptest.NewRequest(verb.method, verb.path, strings.NewReader(verb.body))
			req.Header.Set("Authorization", "Bearer "+piiToken(catalogueAdminID, domain.RoleAgricultor))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
			}
			if len(repo.saved) != 0 {
				t.Errorf("a refused caller wrote %d categories, want 0", len(repo.saved))
			}
		})
	}
}

func TestCatalogueWritesRequireAuthentication(t *testing.T) {
	t.Parallel()

	router, repo := newCatalogueRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(`{"name":"Granos"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 0 {
		t.Errorf("an unauthenticated caller wrote %d categories, want 0", len(repo.saved))
	}
}

func TestCatalogueAdminLifecycleReachesTheDomain(t *testing.T) {
	t.Parallel()

	router, repo := newCatalogueRouter(t)

	created := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(`{"name":"Granos","main_category":"granos"}`))
	created.Header.Set("Authorization", "Bearer "+piiToken(catalogueAdminID, domain.RoleAdmin))
	created.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, created)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 1 {
		t.Fatalf("create wrote %d categories, want 1", len(repo.saved))
	}
	if repo.saved[0].Name != "Granos" || repo.saved[0].MainCategory != "granos" {
		t.Errorf("created = %+v, want the requested name and main category", repo.saved[0])
	}

	newID := repo.saved[0].ID

	status := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/categories/"+newID.String()+"/status", strings.NewReader(`{"is_active":false}`))
	status.Header.Set("Authorization", "Bearer "+piiToken(catalogueAdminID, domain.RoleAdmin))
	status.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, status)

	if rr.Code != http.StatusOK {
		t.Fatalf("status toggle = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.saved) != 2 || repo.saved[1].IsActive {
		t.Fatalf("after the toggle the writes are %+v, want the entry deactivated", repo.saved)
	}
}

func TestPublicCatalogueExcludesDeactivatedEntries(t *testing.T) {
	t.Parallel()

	active, err := domain.NewCategory("Frutales")
	if err != nil {
		t.Fatalf("NewCategory: %v", err)
	}
	retired, err := domain.NewCategory("Hortalizas")
	if err != nil {
		t.Fatalf("NewCategory: %v", err)
	}
	retired.Deactivate()

	repo := &catalogueRepo{categories: []domain.Category{*active, *retired}}

	router := NewRouter(
		nil, nil,
		nil,
		nil,
		handler.NewCategoryHandler(usecases.NewCategoryUseCase(repo)),
		nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "Hortalizas") {
		t.Errorf("the public catalogue returned the deactivated entry: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Frutales") {
		t.Errorf("the public catalogue dropped the active entry: %s", rr.Body.String())
	}
}

func TestOfferingDeactivateRouteReachesTheDomain(t *testing.T) {
	t.Parallel()

	offeringRepo := &ownershipOfferingRepo{offering: offeringWithLocation()}
	uc := usecases.NewOfferingUseCase(
		offeringRepo,
		&ownershipUserRepo{user: &domain.User{ID: ownerFarmerID, Role: domain.RoleAgricultor}},
		ownershipClock{},
		ownershipSearch{},
		ownershipInvalidator{},
	)

	router := NewRouter(
		nil, nil,
		handler.NewOfferingHandler(uc, nil),
		nil, nil, nil, nil,
		middleware.NewAuthMiddleware(ownershipJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)

	path := "/api/v1/offerings/" + ownerOfferingID.String() + "/status"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPatch, path, nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous deactivate status = %d, want 401; body = %s", rr.Code, rr.Body.String())
	}

	req := httptest.NewRequest(http.MethodPatch, path, nil)
	req.Header.Set("Authorization", "Bearer "+ownershipToken(intruderID, domain.RoleAgricultor))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("foreign deactivate status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPatch, path, nil)
	req.Header.Set("Authorization", "Bearer "+ownershipToken(ownerFarmerID, domain.RoleAgricultor))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("owner deactivate status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"is_active":false`) {
		t.Errorf("deactivate body = %s, want the entry reported inactive", rr.Body.String())
	}
}

func offeringWithLocation() *domain.Offering {
	offering, err := domain.NewOffering(ownerFarmerID, "Organic Corn", domain.OfferingProduct, time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	offering.ID = ownerOfferingID
	offering.Variety = "Cuzqueño"
	offering.QuantityAvailable = 100
	return offering
}
