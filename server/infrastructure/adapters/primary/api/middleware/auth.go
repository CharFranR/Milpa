package middleware

import (
	"net/http"
	"strings"

	"github.com/CharFranR/Hackaton2026/domain/port/secondary"
	"github.com/CharFranR/Hackaton2026/internal/auth"
)

type AuthMiddleware struct {
	jwt port.JWTProvider
}

func NewAuthMiddleware(jwt port.JWTProvider) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.jwt.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		principal := auth.Principal{
			UserID: claims.UserID,
			Role:   claims.Role,
		}
		ctx := auth.WithPrincipal(r.Context(), principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
