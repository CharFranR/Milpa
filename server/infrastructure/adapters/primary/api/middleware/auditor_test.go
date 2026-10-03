package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestAuditorMiddlewareIsReadOnly(t *testing.T) {
	t.Parallel()

	auditor := domain.RoleAuditor
	farmer := domain.RoleAgricultor
	admin := domain.RoleAdmin

	tests := []struct {
		name     string
		method   string
		role     *domain.RoleOptions
		wantNext bool
	}{
		{name: "auditor reads with GET", method: http.MethodGet, role: &auditor, wantNext: true},
		{name: "auditor reads with HEAD", method: http.MethodHead, role: &auditor, wantNext: true},
		{name: "auditor preflights with OPTIONS", method: http.MethodOptions, role: &auditor, wantNext: true},
		{name: "auditor is refused on POST", method: http.MethodPost, role: &auditor},
		{name: "auditor is refused on PATCH", method: http.MethodPatch, role: &auditor},
		{name: "auditor is refused on PUT", method: http.MethodPut, role: &auditor},
		{name: "auditor is refused on DELETE", method: http.MethodDelete, role: &auditor},
		{name: "a farmer posts", method: http.MethodPost, role: &farmer, wantNext: true},
		{name: "an admin posts", method: http.MethodPost, role: &admin, wantNext: true},
		{name: "an unauthenticated caller is the authentication chain business", method: http.MethodPost, wantNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reached := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/api/v1/offerings", nil)
			if tt.role != nil {
				ctx := auth.WithPrincipal(req.Context(), auth.Principal{UserID: uuid.New(), Role: *tt.role})
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			NewAuditorMiddleware().CheckReadOnly(next).ServeHTTP(rr, req)

			if tt.wantNext {
				if !reached {
					t.Fatalf("the handler was not reached, status = %d", rr.Code)
				}
				return
			}
			if reached {
				t.Fatal("a read-only auditor reached the handler")
			}
			if rr.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}
