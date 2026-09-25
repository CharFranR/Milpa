package middleware

import (
	"net/http"
	"strings"

	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api/httpx"
	"milpa/internal/auth"

	"github.com/gorilla/websocket"
)

type AuthMiddleware struct {
	jwt port.JWTProvider
}

func NewAuthMiddleware(jwt port.JWTProvider) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeUnauthorized(w, "missing or invalid authorization header")
			return
		}

		m.serveAuthenticated(w, r, token, next)
	})
}

func (m *AuthMiddleware) AuthenticateWebSocket(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := subprotocolToken(r)
		if token == "" {
			token = bearerToken(r.Header.Get("Authorization"))
		}
		if token == "" {
			writeUnauthorized(w, "missing or invalid websocket credentials")
			return
		}

		m.serveAuthenticated(w, r, token, next)
	})
}

func (m *AuthMiddleware) serveAuthenticated(w http.ResponseWriter, r *http.Request, token string, next http.Handler) {
	claims, err := m.jwt.ValidateToken(token)
	if err != nil {
		writeUnauthorized(w, "invalid or expired token")
		return
	}

	principal := auth.Principal{
		UserID: claims.UserID,
		Role:   claims.Role,
	}
	ctx := auth.WithPrincipal(r.Context(), principal)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}

func subprotocolToken(r *http.Request) string {
	for _, protocol := range websocket.Subprotocols(r) {
		if token, found := strings.CutPrefix(protocol, auth.WebSocketTokenPrefix); found && token != "" {
			return token
		}
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	httpx.RespondError(w, http.StatusUnauthorized, message)
}
