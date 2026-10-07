package middleware

import (
	"net/http"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/httpx"
	"milpa/internal/auth"
)

type AdminMiddleware struct{}

func NewAdminMiddleware() *AdminMiddleware {
	return &AdminMiddleware{}
}

func (m *AdminMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.FromContext(r.Context())
		if !ok {
			httpx.RespondError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}
		if principal.Role != domain.RoleAdmin {
			httpx.RespondError(w, http.StatusForbidden, "an administrator role is required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
