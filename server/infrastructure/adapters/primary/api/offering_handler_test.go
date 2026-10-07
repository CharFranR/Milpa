package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api/handler"
)

// visibilityOfferingUC is the transport-level recorder for the include_hidden
// parsing. The policy behind the flag (who may see hidden offerings) is covered
// by the use-case tests; here only the flag decoding matters, so the stub
// answers an empty page and remembers what it was asked.
type visibilityOfferingUC struct {
	includeHidden bool
	called        bool
}

func (s *visibilityOfferingUC) CreateOffering(ctx context.Context, req dto.CreateOfferingRequest) (*dto.OfferingDTO, error) {
	return nil, nil
}

func (s *visibilityOfferingUC) GetByID(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	return nil, nil
}

func (s *visibilityOfferingUC) GetByUserID(ctx context.Context, userID uuid.UUID, includeHidden bool) ([]*dto.OfferingDTO, error) {
	s.called = true
	s.includeHidden = includeHidden
	return []*dto.OfferingDTO{}, nil
}

func (s *visibilityOfferingUC) UpdateOffering(ctx context.Context, id uuid.UUID, req dto.UpdateOfferingRequest) error {
	return nil
}

func (s *visibilityOfferingUC) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (s *visibilityOfferingUC) DeactivateOffering(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	return nil, nil
}

func (s *visibilityOfferingUC) RenewOffering(ctx context.Context, id uuid.UUID, req dto.RenewOfferingRequest) (*dto.OfferingDTO, error) {
	return nil, nil
}

var _ primary.OfferingUseCase = (*visibilityOfferingUC)(nil)

// TestOfferingHandlerIncludeHiddenValues pins the query grammar: the flag takes
// true/1 and false/0 (or nothing), and anything else is a 400. Explicit false
// values are accepted precisely because a client that serializes a boolean
// always sends one, and rejecting "false" punished correct clients.
func TestOfferingHandlerIncludeHiddenValues(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name       string
		value      string
		wantStatus int
		wantHidden bool
		forwarded  bool
	}{
		{name: "omitted", value: "", wantStatus: http.StatusOK},
		{name: "false", value: "false", wantStatus: http.StatusOK},
		{name: "zero", value: "0", wantStatus: http.StatusOK},
		{name: "true", value: "true", wantStatus: http.StatusOK, wantHidden: true},
		{name: "one", value: "1", wantStatus: http.StatusOK, wantHidden: true},
		{name: "anything else is rejected", value: "yes", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &visibilityOfferingUC{}
			h := handler.NewOfferingHandler(uc, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/offerings/?user_id="+userID.String()+"&include_hidden="+tt.value, nil)
			rr := httptest.NewRecorder()

			h.GetByUserID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			wantForwarded := tt.wantStatus == http.StatusOK
			if uc.called != wantForwarded {
				t.Fatalf("use case called = %v, want %v", uc.called, wantForwarded)
			}
			if wantForwarded && uc.includeHidden != tt.wantHidden {
				t.Errorf("includeHidden = %v, want %v", uc.includeHidden, tt.wantHidden)
			}
		})
	}
}
