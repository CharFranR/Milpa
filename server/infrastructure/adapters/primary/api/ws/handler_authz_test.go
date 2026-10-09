package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

type stubConversationUC struct {
	err   error
	gotID uuid.UUID
}

func (s *stubConversationUC) CreateConversation(ctx context.Context, req dto.CreateConversationDTO) (*dto.ConversationDTO, error) {
	return nil, nil
}

func (s *stubConversationUC) ListConversations(ctx context.Context) (*[]dto.ConversationDTO, error) {
	return nil, nil
}

func (s *stubConversationUC) GetConversation(ctx context.Context, id uuid.UUID) (*dto.ConversationDTO, error) {
	s.gotID = id
	return nil, s.err
}

func (s *stubConversationUC) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	return nil
}

func wsRequest(t *testing.T, conversationID string) *http.Request {
	return wsRequestWithRole(t, conversationID, domain.RolePending)
}

func wsRequestWithRole(t *testing.T, conversationID string, role domain.RoleOptions) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/"+conversationID, nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: uuid.New(), Role: role}))

	if conversationID != "" {
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("conversationID", conversationID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	}

	return req
}

func TestWSHandlerAuthorizationStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		conversationID string
		wantStatus     int
	}{
		{name: "not a uuid", conversationID: "nope", wantStatus: http.StatusBadRequest},
		{name: "forbidden", conversationID: uuid.NewString(), wantStatus: http.StatusForbidden},
		{name: "not found", conversationID: uuid.NewString(), wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &stubConversationUC{}
			switch tt.wantStatus {
			case http.StatusForbidden:
				uc.err = domain.ErrForbidden
			case http.StatusNotFound:
				uc.err = domain.ErrNotFound
			}

			h := NewHandler(NewHub(), nil, uc)
			rr := httptest.NewRecorder()

			h.WSHandler(rr, wsRequest(t, tt.conversationID))

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus != http.StatusBadRequest && uc.gotID.String() != tt.conversationID {
				t.Fatalf("use case received id %v, want %v", uc.gotID, tt.conversationID)
			}
		})
	}
}

func TestWSHandlerRefusesAuditor(t *testing.T) {
	t.Parallel()

	uc := &stubConversationUC{}
	h := NewHandler(NewHub(), nil, uc)
	rr := httptest.NewRecorder()

	h.WSHandler(rr, wsRequestWithRole(t, uuid.NewString(), domain.RoleAuditor))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	if uc.gotID != uuid.Nil {
		t.Fatalf("auditor reached the conversation use case with id %v", uc.gotID)
	}
}

func TestWSHandlerPassesNonAuditorPastTheRoleGate(t *testing.T) {
	t.Parallel()

	uc := &stubConversationUC{}
	h := NewHandler(NewHub(), nil, uc)
	rr := httptest.NewRecorder()

	h.WSHandler(rr, wsRequestWithRole(t, "nope", domain.RoleAgricultor))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}
