package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/internal/auth"
)

type stubJWT struct{}

func (stubJWT) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	return "", nil
}

func (stubJWT) ValidateToken(token string) (*port.JWTClaims, error) {
	if token != "valid-token" {
		return nil, errors.New("invalid")
	}
	return &port.JWTClaims{UserID: uuid.New(), Role: domain.RoleMIPYME}, nil
}

func newWSServer(t *testing.T) *httptest.Server {
	t.Helper()

	mw := middleware.NewAuthMiddleware(stubJWT{})

	handler := mw.AuthenticateWebSocket(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := auth.RequirePrincipal(r.Context())
		if err != nil {
			http.Error(w, "no principal", http.StatusInternalServerError)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if err := conn.WriteJSON(map[string]string{"user_id": principal.UserID.String()}); err != nil {
			t.Errorf("writejson: %v", err)
		}
	}))

	return httptest.NewServer(handler)
}

func wsURL(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func TestHandshakeWithSubprotocolToken(t *testing.T) {
	server := newWSServer(t)
	defer server.Close()

	dialer := websocket.Dialer{
		Subprotocols: []string{auth.WebSocketProtocol, auth.WebSocketTokenPrefix + "valid-token"},
	}

	conn, resp, err := dialer.Dial(wsURL(server), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", resp.StatusCode)
	}

	got := resp.Header.Get("Sec-WebSocket-Protocol")
	if got != auth.WebSocketProtocol {
		t.Fatalf("negotiated subprotocol = %q, want %q", got, auth.WebSocketProtocol)
	}
	if strings.Contains(got, "valid-token") {
		t.Fatalf("token leaked into negotiated subprotocol: %q", got)
	}

	var payload map[string]string
	if err := conn.ReadJSON(&payload); err != nil {
		t.Fatalf("readjson: %v", err)
	}
	if _, err := uuid.Parse(payload["user_id"]); err != nil {
		t.Fatalf("principal missing from handler ctx: %v", err)
	}
}

func TestHandshakeRejectsInvalidToken(t *testing.T) {
	server := newWSServer(t)
	defer server.Close()

	dialer := websocket.Dialer{
		Subprotocols: []string{auth.WebSocketProtocol, auth.WebSocketTokenPrefix + "bad-token"},
	}

	conn, resp, err := dialer.Dial(wsURL(server), nil)
	if err == nil {
		conn.Close()
		t.Fatal("dial succeeded, want failure")
	}
	if resp == nil {
		t.Fatalf("no response from handshake: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid or expired token" {
		t.Fatalf("error = %q", body["error"])
	}
}

func TestHandshakeRejectsMissingCredentials(t *testing.T) {
	server := newWSServer(t)
	defer server.Close()

	dialer := websocket.Dialer{Subprotocols: []string{auth.WebSocketProtocol}}

	_, resp, err := dialer.Dial(wsURL(server), nil)
	if err == nil {
		t.Fatal("dial succeeded, want failure")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestHandshakeAcceptsAuthorizationHeaderFallback(t *testing.T) {
	server := newWSServer(t)
	defer server.Close()

	dialer := websocket.Dialer{Subprotocols: []string{auth.WebSocketProtocol}}
	header := http.Header{"Authorization": []string{"Bearer valid-token"}}

	conn, resp, err := dialer.Dial(wsURL(server), header)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", resp.StatusCode)
	}
	if got := resp.Header.Get("Sec-WebSocket-Protocol"); got != auth.WebSocketProtocol {
		t.Fatalf("negotiated subprotocol = %q, want %q", got, auth.WebSocketProtocol)
	}
}
