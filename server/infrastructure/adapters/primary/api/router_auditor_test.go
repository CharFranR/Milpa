package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
)

func TestAuditorIsRefusedOnTheWiredRoutes(t *testing.T) {
	t.Parallel()

	router, repo := newOwnershipRouter(t)
	auditorID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		role       domain.RoleOptions
		wantStatus int
	}{
		{
			name:       "auditor cannot POST an offering",
			method:     http.MethodPost,
			path:       "/api/v1/offerings/",
			body:       `{}`,
			role:       domain.RoleAuditor,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "auditor cannot delete an offering",
			method:     http.MethodPatch,
			path:       "/api/v1/offerings/" + ownerOfferingID.String(),
			body:       `{"name":"Renamed"}`,
			role:       domain.RoleAuditor,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "a farmer is not refused by the auditor middleware",
			method:     http.MethodPost,
			path:       "/api/v1/offerings/",
			body:       ``,
			role:       domain.RoleAgricultor,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+ownershipToken(auditorID, tt.role))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}

	if len(repo.deleted) != 0 {
		t.Errorf("the auditor deleted %v, want nothing", repo.deleted)
	}
}
