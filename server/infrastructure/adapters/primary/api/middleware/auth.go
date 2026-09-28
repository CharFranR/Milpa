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

// AuthenticateOptional attaches a principal when the request carries usable
// credentials and otherwise lets it through untouched.
//
// It exists for the public marketplace reads, which are reachable without a
// token but have to answer the owner and an admin with more than an anonymous
// visitor gets. Requiring authentication there would close the marketplace;
// ignoring the header would make the owner unable to see their own contact card.
// Missing or unusable credentials degrade to anonymous rather than failing: a
// stale token in a browser must not turn a public page into a 401.
//
// It never upgrades a request, only annotates it. A route that requires
// authentication must use Authenticate.
func (m *AuthMiddleware) AuthenticateOptional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwt.ValidateToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		principal := auth.Principal{
			UserID: claims.UserID,
			Role:   claims.Role,
		}
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
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
