package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

type stubUnitRepo struct {
	units []domain.UnitOfMeasure
}

func (r *stubUnitRepo) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return r.units, nil
}

func newUnitsRouter(t *testing.T) http.Handler {
	t.Helper()

	kilo := uuid.MustParse("c2dba60d-a701-4e93-990a-f2fb814ad114")
	uc := usecases.NewUnitOfMeasureUseCase(&stubUnitRepo{units: []domain.UnitOfMeasure{
		{ID: kilo, Code: "kg", Name: "Kilogramo", IsActive: true},
	}})

	return NewRouter(
		nil, nil, nil, nil, nil, nil, nil,
		middleware.NewAuthMiddleware(piiJWT{}), middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		handler.NewUnitOfMeasureHandler(uc),
	)
}

func TestUnitsOfMeasureIsPublic(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	newUnitsRouter(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/units-of-measure", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	var body struct {
		Data []struct {
			ID   uuid.UUID `json:"id"`
			Code string    `json:"code"`
			Name string    `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not the {data: ...} envelope: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].Code != "kg" {
		t.Fatalf("data = %+v, want the single kg unit", body.Data)
	}
}