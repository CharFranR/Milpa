package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "milpa/domain/entities"
)

func TestRespondNoContentHasNoBody(t *testing.T) {
	recorder := httptest.NewRecorder()

	Respond(recorder, http.StatusNoContent, nil)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
}

func TestStatusCodeMappings(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "phone number required is a validation error", err: domain.ErrPhoneNumberRequired, want: http.StatusBadRequest},
		{name: "wrapped phone number required keeps the mapping", err: fmt.Errorf("user.Save: %w", domain.ErrPhoneNumberRequired), want: http.StatusBadRequest},
		{name: "valid role required is a validation error", err: domain.ErrValidRoleRequired, want: http.StatusBadRequest},
		{name: "review requires a transaction", err: domain.ErrTransactionRequired, want: http.StatusBadRequest},
		{name: "review target party mismatch is a validation error", err: domain.ErrReviewTargetPartyMismatch, want: http.StatusBadRequest},
		{name: "invalid review target type is a validation error", err: domain.ErrInvalidReviewTargetType, want: http.StatusBadRequest},
		{name: "message content required is a validation error", err: domain.ErrContentMessageRequired, want: http.StatusBadRequest},
		{name: "review already exists is a conflict", err: domain.ErrReviewAlreadyExists, want: http.StatusConflict},
		{name: "transaction not completed is a conflict", err: domain.ErrTransactionNotCompleted, want: http.StatusConflict},
		{name: "liquidation not open is a conflict", err: domain.ErrLiquidationNotOpen, want: http.StatusConflict},
		{name: "liquidation cannot assign is a conflict", err: domain.ErrLiquidationCannotAssign, want: http.StatusConflict},
		{name: "user not found", err: domain.ErrUserNotFound, want: http.StatusNotFound},
		{name: "inventory not found", err: domain.ErrInventoryNotFound, want: http.StatusNotFound},
		{name: "suspended user is forbidden", err: domain.ErrUserSuspended, want: http.StatusForbidden},
		{name: "unknown error stays internal", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCode(tt.err); got != tt.want {
				t.Errorf("StatusCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
