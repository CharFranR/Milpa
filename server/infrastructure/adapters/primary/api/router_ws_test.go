package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
)

type stubJWT struct{}

func (stubJWT) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	return "", nil
}

func (stubJWT) ValidateToken(token string) (*port.JWTClaims, error) {
	if token != "valid-token" {
		return nil, errors.New("invalid or expired token")
	}
	return &port.JWTClaims{UserID: uuid.New(), Role: domain.RoleMIPYME}, nil
}

type stubUserRepo struct{}

func (stubUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return &domain.User{}, nil
}

func (stubUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (stubUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (stubUserRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (stubUserRepo) Save(ctx context.Context, user *domain.User) (string, error) {
	return "", nil
}

func (stubUserRepo) Update(ctx context.Context, user *domain.User) error {
	return nil
}

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

func newTestRouter(t *testing.T, convUC *stubConversationUC) http.Handler {
	t.Helper()

	chat := ws.NewHandler(ws.NewHub(), nil, convUC)
	authMW := middleware.NewAuthMiddleware(stubJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	return NewRouter(nil, nil, nil, nil, nil, nil, nil, authMW, suspensionMW, nil, nil, nil, nil, nil, nil, chat)
}

func TestChatWebSocketRouteRequiresToken(t *testing.T) {
	t.Parallel()

	convUC := &stubConversationUC{err: domain.ErrForbidden}
	router := newTestRouter(t, convUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/"+uuid.NewString(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (route missing or auth not attached); body = %s", rr.Code, rr.Body.String())
	}
}

func TestChatWebSocketRouteExtractsConversationID(t *testing.T) {
	t.Parallel()

	convUC := &stubConversationUC{err: domain.ErrForbidden}
	router := newTestRouter(t, convUC)

	conversationID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/"+conversationID, nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}
	if convUC.gotID.String() != conversationID {
		t.Fatalf("use case received id %v, want %v — route param name does not match chi.URLParam", convUC.gotID, conversationID)
	}
}

func TestChatWebSocketRouteUnknownPathIs404(t *testing.T) {
	t.Parallel()

	convUC := &stubConversationUC{err: domain.ErrForbidden}
	router := newTestRouter(t, convUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}
