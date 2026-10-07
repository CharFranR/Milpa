package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/internal/auth"
)

type stubTransactionUseCase struct {
	err              error
	list             []*dto.TransactionDTO
	gotMatchID       uuid.UUID
	gotRequestID     uuid.UUID
	gotTransactionID uuid.UUID
	gotReason        string
}

func (s *stubTransactionUseCase) GetByMatch(ctx context.Context, matchID uuid.UUID) (*dto.TransactionDTO, error) {
	s.gotMatchID = matchID
	if s.err != nil {
		return nil, s.err
	}
	return &dto.TransactionDTO{Status: domain.TransactionMatched}, nil
}

func (s *stubTransactionUseCase) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.TransactionDTO, error) {
	s.gotRequestID = supplyRequestID
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubTransactionUseCase) ConfirmStart(ctx context.Context, transactionID uuid.UUID) error {
	s.gotTransactionID = transactionID
	return s.err
}

func (s *stubTransactionUseCase) ConfirmDelivery(ctx context.Context, transactionID uuid.UUID) error {
	s.gotTransactionID = transactionID
	return s.err
}

func (s *stubTransactionUseCase) Cancel(ctx context.Context, transactionID uuid.UUID, reason string) error {
	s.gotTransactionID = transactionID
	s.gotReason = reason
	return s.err
}

func newTxTestRouter(t *testing.T, uc *stubTransactionUseCase) http.Handler {
	t.Helper()

	authMW := middleware.NewAuthMiddleware(stubJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	r := NewRouter(nil, nil, nil, nil, nil, nil, nil, authMW, suspensionMW, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	RegisterTransactionRoutes(r, handler.NewTransactionHandler(uc), authMW, suspensionMW)
	return r
}

func TestTransactionRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "get by match", method: http.MethodGet, path: "/api/v1/transactions/matches/" + uuid.NewString()},
		{name: "list by request", method: http.MethodGet, path: "/api/v1/transactions/requests/" + uuid.NewString()},
		{name: "confirm start", method: http.MethodPost, path: "/api/v1/transactions/" + uuid.NewString() + "/confirm-start"},
		{name: "confirm delivery", method: http.MethodPost, path: "/api/v1/transactions/" + uuid.NewString() + "/confirm-delivery"},
		{name: "cancel", method: http.MethodPost, path: "/api/v1/transactions/" + uuid.NewString() + "/cancel", body: `{"reason":"x"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTxTestRouter(t, &stubTransactionUseCase{})

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (route missing or auth not attached); body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestTransactionRouteExtractsIdentifiers(t *testing.T) {
	t.Parallel()

	t.Run("get by match extracts match id", func(t *testing.T) {
		t.Parallel()

		uc := &stubTransactionUseCase{}
		router := newTxTestRouter(t, uc)
		matchID := uuid.NewString()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/matches/"+matchID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
		}
		if uc.gotMatchID.String() != matchID {
			t.Fatalf("use case received match id %v, want %v", uc.gotMatchID, matchID)
		}
	})

	t.Run("list by request extracts request id", func(t *testing.T) {
		t.Parallel()

		uc := &stubTransactionUseCase{list: []*dto.TransactionDTO{}}
		router := newTxTestRouter(t, uc)
		requestID := uuid.NewString()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/requests/"+requestID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
		}
		if uc.gotRequestID.String() != requestID {
			t.Fatalf("use case received request id %v, want %v", uc.gotRequestID, requestID)
		}
	})

	t.Run("confirm start extracts transaction id", func(t *testing.T) {
		t.Parallel()

		uc := &stubTransactionUseCase{}
		router := newTxTestRouter(t, uc)
		transactionID := uuid.NewString()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/"+transactionID+"/confirm-start", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
		}
		if uc.gotTransactionID.String() != transactionID {
			t.Fatalf("use case received transaction id %v, want %v", uc.gotTransactionID, transactionID)
		}
	})

	t.Run("confirm delivery extracts transaction id", func(t *testing.T) {
		t.Parallel()

		uc := &stubTransactionUseCase{}
		router := newTxTestRouter(t, uc)
		transactionID := uuid.NewString()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/"+transactionID+"/confirm-delivery", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
		}
		if uc.gotTransactionID.String() != transactionID {
			t.Fatalf("use case received transaction id %v, want %v", uc.gotTransactionID, transactionID)
		}
	})
}

func TestTransactionCancelRouteDecodesReason(t *testing.T) {
	t.Parallel()

	uc := &stubTransactionUseCase{}
	router := newTxTestRouter(t, uc)
	transactionID := uuid.NewString()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/"+transactionID+"/cancel", strings.NewReader(`{"reason":"supplier missed the delivery window"}`))
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if uc.gotReason != "supplier missed the delivery window" {
		t.Fatalf("reason = %q, want the decoded body reason", uc.gotReason)
	}
	if uc.gotTransactionID.String() != transactionID {
		t.Fatalf("use case received transaction id %v, want %v", uc.gotTransactionID, transactionID)
	}
}

func TestTransactionCancelRouteRejectsInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTxTestRouter(t, &stubTransactionUseCase{})
	transactionID := uuid.NewString()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/"+transactionID+"/cancel", strings.NewReader(`not-json`))
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
	}
}

func TestTransactionErrorStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "forbidden maps to 403", err: domain.ErrForbidden, status: http.StatusForbidden},
		{name: "not found maps to 404", err: domain.ErrNotFound, status: http.StatusNotFound},
		{name: "terminal state maps to 409", err: domain.ErrTerminalState, status: http.StatusConflict},
		{name: "invalid transition maps to 409", err: domain.ErrInvalidTransactionTransition, status: http.StatusConflict},
		{name: "already confirmed maps to 409", err: domain.ErrAlreadyConfirmed, status: http.StatusConflict},
		{name: "invalid match status maps to 409", err: domain.ErrInvalidMatchStatus, status: http.StatusConflict},
		{name: "reason required maps to 400", err: domain.ErrReasonRequired, status: http.StatusBadRequest},
		{name: "unauthenticated maps to 401", err: auth.ErrUnauthenticated, status: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTxTestRouter(t, &stubTransactionUseCase{err: tt.err})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/matches/"+uuid.NewString(), nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.status, rr.Body.String())
			}
		})
	}
}

func TestTransactionUnknownPathIs404(t *testing.T) {
	t.Parallel()

	router := newTxTestRouter(t, &stubTransactionUseCase{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}
