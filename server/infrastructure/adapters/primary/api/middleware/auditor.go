package middleware

import (
	"net/http"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/httpx"
	"milpa/internal/auth"
)

type AuditorMiddleware struct{}

func NewAuditorMiddleware() *AuditorMiddleware {
	return &AuditorMiddleware{}
}

// CheckReadOnly runs after the authentication middlewares and refuses the
// state-changing methods for an auditor: the role exists to read the
// marketplace, not to write to it. An unauthenticated caller is not the
// middleware's business, the authentication chain already answered for it.
func (m *AuditorMiddleware) CheckReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.FromContext(r.Context())
		if !ok || principal.Role != domain.RoleAuditor {
			next.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
		default:
			httpx.RespondError(w, http.StatusForbidden, "an auditor has read-only access")
		}
	})
}
