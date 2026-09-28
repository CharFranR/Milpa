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
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
)

// The offering mutation routes are the IDOR this file covers. The stubs below
// are wired to the real use case so the assertions exercise the actual policy
// and not a restatement of it; the unit tests in
// tests/unitary/usecases/offering_ownership_test.go cover the same policy with
// the case matrix.

var (
	ownerOfferingID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ownerFarmerID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	intruderID      = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

type ownershipOfferingRepo struct {
	offering *domain.Offering
	deleted  []uuid.UUID
}

func (r *ownershipOfferingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
	if r.offering == nil || r.offering.ID != id {
		return nil, domain.ErrNotFound
	}
	copied := *r.offering
	return &copied, nil
}

func (r *ownershipOfferingRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Offering, error) {
	return nil, nil
}

func (r *ownershipOfferingRepo) Save(ctx context.Context, offering *domain.Offering) error {
	return nil
}

func (r *ownershipOfferingRepo) Update(ctx context.Context, offering *domain.Offering) error {
	return nil
}

func (r *ownershipOfferingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	return nil
}

var _ port.OfferingRepository = (*ownershipOfferingRepo)(nil)

type ownershipUserRepo struct {
	user *domain.User
}

func (r *ownershipUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if r.user == nil || r.user.ID != id {
		return nil, domain.ErrNotFound
	}
	return r.user, nil
}

func (r *ownershipUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (r *ownershipUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (r *ownershipUserRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (r *ownershipUserRepo) Save(ctx context.Context, user *domain.User) (string, error) {
	return "", nil
}

func (r *ownershipUserRepo) Update(ctx context.Context, user *domain.User) error {
	return nil
}

var _ port.UserRepository = (*ownershipUserRepo)(nil)

type ownershipClock struct{}

func (ownershipClock) Now() time.Time { return time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC) }

type ownershipSearch struct{}

func (ownershipSearch) Search(ctx context.Context, query *dto.SearchQuery) (*dto.SearchResponse, error) {
	return &dto.SearchResponse{}, nil
}

func (ownershipSearch) Index(ctx context.Context, p *dto.IndexOfferingRequest) error { return nil }
func (ownershipSearch) Update(ctx context.Context, id string, p *dto.IndexOfferingRequest) error {
	return nil
}
func (ownershipSearch) Delete(ctx context.Context, id string) error { return nil }

type ownershipInvalidator struct{}

func (ownershipInvalidator) InvalidateAll(ctx context.Context) error { return nil }

// ownershipJWT encodes the caller in the token as "<userID>:<role>", for the
// same reason piiJWT does: the stock stubJWT hands every token one identity.
type ownershipJWT struct{}

func (ownershipJWT) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	return ownershipToken(userID, role), nil
}

func (ownershipJWT) ValidateToken(token string) (*port.JWTClaims, error) {
	return piiJWT{}.ValidateToken(token)
}

func ownershipToken(userID uuid.UUID, role domain.RoleOptions) string {
	return piiToken(userID, role)
}

func newOwnershipRouter(t *testing.T) (http.Handler, *ownershipOfferingRepo) {
	t.Helper()

	offering, err := domain.NewOffering(ownerFarmerID, "Organic Corn", domain.OfferingProduct, time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewOffering: %v", err)
	}
	offering.ID = ownerOfferingID

	repo := &ownershipOfferingRepo{offering: offering}
	uc := usecases.NewOfferingUseCase(
		repo,
		&ownershipUserRepo{user: &domain.User{ID: ownerFarmerID, Role: domain.RoleProvider}},
		ownershipClock{},
		ownershipSearch{},
		ownershipInvalidator{},
	)

	chat := ws.NewHandler(ws.NewHub(), nil, &stubConversationUC{})
	authMW := middleware.NewAuthMiddleware(ownershipJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	router := NewRouter(
		nil, nil,
		handler.NewOfferingHandler(uc, nil),
		nil, nil, nil, nil,
		authMW, suspensionMW,
		nil, nil, nil, nil, nil, nil, chat,
		nil, nil, nil, nil,
	)

	return router, repo
}

// TestOfferingMutationRefusesForeignUser is the IDOR at the transport boundary.
// chi keeps the LAST handler registered for a method+pattern, so
// PATCH /api/v1/offerings/{id} reaches DeleteOffering, not Update; the route
// registration is left as it is and both use cases are fixed in their own tests.
func TestOfferingMutationRefusesForeignUser(t *testing.T) {
	t.Parallel()

	router, repo := newOwnershipRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/offerings/"+ownerOfferingID.String(), strings.NewReader(`{"name":"Hijacked"}`))
	req.Header.Set("Authorization", "Bearer "+ownershipToken(intruderID, domain.RoleProvider))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.deleted) != 0 {
		t.Errorf("a refused caller deleted %v, want nothing", repo.deleted)
	}
}

func TestOfferingMutationAllowsOwner(t *testing.T) {
	t.Parallel()

	router, repo := newOwnershipRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/offerings/"+ownerOfferingID.String(), strings.NewReader(`{"name":"Renamed"}`))
	req.Header.Set("Authorization", "Bearer "+ownershipToken(ownerFarmerID, domain.RoleProvider))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for the owner; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != ownerOfferingID {
		t.Errorf("deleted = %v, want [%v]", repo.deleted, ownerOfferingID)
	}
}

func TestOfferingMutationRequiresAuthentication(t *testing.T) {
	t.Parallel()

	router, repo := newOwnershipRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/offerings/"+ownerOfferingID.String(), strings.NewReader(`{"name":"Hijacked"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.deleted) != 0 {
		t.Errorf("an unauthenticated caller deleted %v, want nothing", repo.deleted)
	}
}

func TestOfferingMutationUnknownIDIsNotFound(t *testing.T) {
	t.Parallel()

	router, repo := newOwnershipRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/offerings/"+uuid.NewString(), strings.NewReader(`{"name":"Whatever"}`))
	req.Header.Set("Authorization", "Bearer "+ownershipToken(ownerFarmerID, domain.RoleProvider))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rr.Code, rr.Body.String())
	}
	if len(repo.deleted) != 0 {
		t.Errorf("an unknown id deleted %v, want nothing", repo.deleted)
	}
}
