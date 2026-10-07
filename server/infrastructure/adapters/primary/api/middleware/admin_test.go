package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestAdminMiddlewareRequiresTheAdminRole(t *testing.T) {
	t.Parallel()

	admin := domain.RoleAdmin
	farmer := domain.RoleAgricultor
	auditor := domain.RoleAuditor

	tests := []struct {
		name     string
		method   string
		role     *domain.RoleOptions
		wantNext bool
		wantCode int
	}{
		{name: "an admin posts", method: http.MethodPost, role: &admin, wantNext: true},
		{name: "an admin patches", method: http.MethodPatch, role: &admin, wantNext: true},
		{name: "a farmer is refused on POST", method: http.MethodPost, role: &farmer, wantCode: http.StatusForbidden},
		{name: "a farmer is refused on PATCH", method: http.MethodPatch, role: &farmer, wantCode: http.StatusForbidden},
		{name: "an auditor is refused on a mutation", method: http.MethodPatch, role: &auditor, wantCode: http.StatusForbidden},
		{name: "an unauthenticated caller is refused", method: http.MethodPost, wantCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reached := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/api/v1/admin/units-of-measure", nil)
			if tt.role != nil {
				ctx := auth.WithPrincipal(req.Context(), auth.Principal{UserID: uuid.New(), Role: *tt.role})
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			NewAdminMiddleware().RequireAdmin(next).ServeHTTP(rr, req)

			if tt.wantNext {
				if !reached {
					t.Fatalf("the handler was not reached, status = %d", rr.Code)
				}
				return
			}
			if reached {
				t.Fatal("a non-admin caller reached the handler")
			}
			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d; body = %s", rr.Code, tt.wantCode, rr.Body.String())
			}
		})
	}
}
