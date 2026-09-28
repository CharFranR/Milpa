package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
)

type stubSupplyRequestUC struct {
	createErr     error
	updateErr     error
	cancelErr     error
	list          []*dto.SupplyRequestDTO
	gotID         uuid.UUID
	gotSupplierID uuid.UUID
}

func (s *stubSupplyRequestUC) Create(ctx context.Context, req dto.SupplyRequestDTO) (*dto.SupplyRequestDTO, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	return &dto.SupplyRequestDTO{ProductName: req.ProductName}, nil
}

func (s *stubSupplyRequestUC) Update(ctx context.Context, id uuid.UUID, req dto.SupplyGeneralUpdateDTO) error {
	s.gotID = id
	return s.updateErr
}

func (s *stubSupplyRequestUC) UpdateAmounts(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateAmountsDTO) error {
	s.gotID = id
	return s.updateErr
}

func (s *stubSupplyRequestUC) UpdateDeadlines(ctx context.Context, id uuid.UUID, req dto.SupplyUpdateTimeDTO) error {
	s.gotID = id
	return s.updateErr
}

func (s *stubSupplyRequestUC) Cancel(ctx context.Context, id uuid.UUID) error {
	s.gotID = id
	return s.cancelErr
}

func (s *stubSupplyRequestUC) Expire(ctx context.Context, id uuid.UUID) error {
	s.gotID = id
	return s.cancelErr
}

func (s *stubSupplyRequestUC) GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyRequestDTO, error) {
	s.gotID = id
	return &dto.SupplyRequestDTO{}, nil
}

func (s *stubSupplyRequestUC) List(ctx context.Context) ([]*dto.SupplyRequestDTO, error) {
	return s.list, nil
}

func (s *stubSupplyRequestUC) ListAvailable(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyRequestDTO, error) {
	s.gotSupplierID = supplierID
	return s.list, nil
}

type stubSupplyOfferUC struct {
	createErr        error
	updateErr        error
	withdrawErr      error
	listByRequestErr error
	gotID            uuid.UUID
	gotRequestID     uuid.UUID
	gotSupplierID    uuid.UUID
}

func (s *stubSupplyOfferUC) Create(ctx context.Context, req dto.SupplyOfferDTO) (*dto.SupplyOfferDTO, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	return &dto.SupplyOfferDTO{TotalAmount: req.TotalAmount}, nil
}

func (s *stubSupplyOfferUC) Update(ctx context.Context, id uuid.UUID, req dto.SupplyOfferUpdateDTO) error {
	s.gotID = id
	return s.updateErr
}

func (s *stubSupplyOfferUC) Withdraw(ctx context.Context, id uuid.UUID) error {
	s.gotID = id
	return s.withdrawErr
}

func (s *stubSupplyOfferUC) GetByID(ctx context.Context, id uuid.UUID) (*dto.SupplyOfferDTO, error) {
	s.gotID = id
	return &dto.SupplyOfferDTO{}, nil
}

func (s *stubSupplyOfferUC) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.SupplyOfferDTO, error) {
	s.gotRequestID = supplyRequestID
	if s.listByRequestErr != nil {
		return nil, s.listByRequestErr
	}
	return nil, nil
}

func (s *stubSupplyOfferUC) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]*dto.SupplyOfferDTO, error) {
	s.gotSupplierID = supplierID
	return nil, nil
}

func newSupplyTestRouter(t *testing.T, requestUC *stubSupplyRequestUC, offerUC *stubSupplyOfferUC) http.Handler {
	t.Helper()

	chat := ws.NewHandler(ws.NewHub(), nil, &stubConversationUC{})
	authMW := middleware.NewAuthMiddleware(stubJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	return NewRouter(
		nil, nil, nil, nil, nil, nil, nil, authMW, suspensionMW, nil, nil, nil, nil, nil, nil, chat,
		handler.NewSupplyRequestHandler(requestUC), handler.NewSupplyOfferHandler(offerUC),
		nil, nil,
	)
}

func supplyAuthenticatedRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestSupplyRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, &stubSupplyOfferUC{})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list supply requests", method: http.MethodGet, path: "/api/v1/supply-requests/"},
		{name: "create supply request", method: http.MethodPost, path: "/api/v1/supply-requests/"},
		{name: "list available supply requests", method: http.MethodGet, path: "/api/v1/supply-requests/available"},
		{name: "get supply request", method: http.MethodGet, path: "/api/v1/supply-requests/" + uuid.NewString()},
		{name: "update supply request", method: http.MethodPatch, path: "/api/v1/supply-requests/" + uuid.NewString()},
		{name: "update supply request amounts", method: http.MethodPatch, path: "/api/v1/supply-requests/" + uuid.NewString() + "/amounts"},
		{name: "update supply request deadlines", method: http.MethodPatch, path: "/api/v1/supply-requests/" + uuid.NewString() + "/deadlines"},
		{name: "cancel supply request", method: http.MethodPost, path: "/api/v1/supply-requests/" + uuid.NewString() + "/cancel"},
		{name: "expire supply request", method: http.MethodPost, path: "/api/v1/supply-requests/" + uuid.NewString() + "/expire"},
		{name: "list supply offers", method: http.MethodGet, path: "/api/v1/supply-offers/"},
		{name: "create supply offer", method: http.MethodPost, path: "/api/v1/supply-offers/"},
		{name: "list offers by request", method: http.MethodGet, path: "/api/v1/supply-offers/requests/" + uuid.NewString()},
		{name: "get supply offer", method: http.MethodGet, path: "/api/v1/supply-offers/" + uuid.NewString()},
		{name: "update supply offer", method: http.MethodPatch, path: "/api/v1/supply-offers/" + uuid.NewString()},
		{name: "withdraw supply offer", method: http.MethodPost, path: "/api/v1/supply-offers/" + uuid.NewString() + "/withdraw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestSupplyRequestCreateReturnsCreated(t *testing.T) {
	t.Parallel()

	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, &stubSupplyOfferUC{})
	req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-requests/", `{"product_name":"Rice"}`)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rr.Code, rr.Body.String())
	}
}

func TestSupplyRequestListAvailableUsesPrincipal(t *testing.T) {
	t.Parallel()

	uc := &stubSupplyRequestUC{}
	router := newSupplyTestRouter(t, uc, &stubSupplyOfferUC{})
	req := supplyAuthenticatedRequest(t, http.MethodGet, "/api/v1/supply-requests/available", "")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if uc.gotSupplierID == uuid.Nil {
		t.Fatal("expected the authenticated principal as supplier id, got nil UUID")
	}
}

func TestSupplyRequestCancelConflictReturns409(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cancelErr error
	}{
		{name: "active match", cancelErr: primary.ErrActiveMatch},
		{name: "not open", cancelErr: domain.ErrInvalidRequestStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestID := uuid.NewString()
			uc := &stubSupplyRequestUC{cancelErr: tt.cancelErr}
			router := newSupplyTestRouter(t, uc, &stubSupplyOfferUC{})
			req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-requests/"+requestID+"/cancel", "")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409; body = %s", rr.Code, rr.Body.String())
			}
			if uc.gotID.String() != requestID {
				t.Fatalf("use case received id %v, want %v", uc.gotID, requestID)
			}
		})
	}
}

func TestSupplyRequestUpdateForbiddenReturns403(t *testing.T) {
	t.Parallel()

	uc := &stubSupplyRequestUC{updateErr: domain.ErrForbidden}
	router := newSupplyTestRouter(t, uc, &stubSupplyOfferUC{})
	req := supplyAuthenticatedRequest(t, http.MethodPatch, "/api/v1/supply-requests/"+uuid.NewString(), `{"product_name":"Maize"}`)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}
}

func TestSupplyRequestCreateInvalidBodyReturns400(t *testing.T) {
	t.Parallel()

	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, &stubSupplyOfferUC{})
	req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-requests/", `{"product_name":`)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
	}
}

func TestSupplyRequestCreateInvalidInputReturns400(t *testing.T) {
	t.Parallel()

	uc := &stubSupplyRequestUC{createErr: fmt.Errorf("%w: product name is required", domain.ErrInvalidInput)}
	router := newSupplyTestRouter(t, uc, &stubSupplyOfferUC{})
	req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-requests/", `{}`)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
	}
}

func TestSupplyOfferCreateConflictReturns409(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		createErr error
	}{
		{name: "duplicate offer", createErr: domain.ErrDuplicate},
		{name: "active match single provider", createErr: primary.ErrActiveMatch},
		{name: "exceeds remaining amount", createErr: domain.ErrInsufficientAmount},
		{name: "request not open", createErr: domain.ErrInvalidRequestStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &stubSupplyOfferUC{createErr: tt.createErr}
			router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, uc)
			req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-offers/", `{"total_amount":20}`)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409; body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestSupplyOfferWithdrawConflictReturns409(t *testing.T) {
	t.Parallel()

	offerID := uuid.NewString()
	uc := &stubSupplyOfferUC{withdrawErr: domain.ErrInvalidOfferStatus}
	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, uc)
	req := supplyAuthenticatedRequest(t, http.MethodPost, "/api/v1/supply-offers/"+offerID+"/withdraw", "")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rr.Code, rr.Body.String())
	}
	if uc.gotID.String() != offerID {
		t.Fatalf("use case received id %v, want %v", uc.gotID, offerID)
	}
}

func TestSupplyOfferListByRequestExtractsRequestID(t *testing.T) {
	t.Parallel()

	requestID := uuid.New()
	uc := &stubSupplyOfferUC{}
	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, uc)
	req := supplyAuthenticatedRequest(t, http.MethodGet, "/api/v1/supply-offers/requests/"+requestID.String(), "")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if uc.gotRequestID != requestID {
		t.Fatalf("use case received request id %v, want %v", uc.gotRequestID, requestID)
	}
}

func TestSupplyOfferListBySupplierUsesPrincipal(t *testing.T) {
	t.Parallel()

	uc := &stubSupplyOfferUC{}
	router := newSupplyTestRouter(t, &stubSupplyRequestUC{}, uc)
	req := supplyAuthenticatedRequest(t, http.MethodGet, "/api/v1/supply-offers/", "")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if uc.gotSupplierID == uuid.Nil {
		t.Fatal("expected the authenticated principal as supplier id, got nil UUID")
	}
}
