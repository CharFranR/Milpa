package middleware

import (
	"net/http"

	"milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type SuspensionMiddleware struct {
	userRepo port.UserRepository
}

func NewSuspensionMiddleware(userRepo port.UserRepository) *SuspensionMiddleware {
	return &SuspensionMiddleware{userRepo: userRepo}
}

func (m *SuspensionMiddleware) CheckSuspension(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.FromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.userRepo.FindByID(r.Context(), principal.UserID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if user.IsSuspended() {
			http.Error(w, `{"error":"your account has been suspended"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
