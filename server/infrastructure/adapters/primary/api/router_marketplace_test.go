package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"milpa/aplication/dto"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
)

// countingFuzzyUC counts calls so this test proves the request reached the
// search listing, not just that some handler answered.
type countingFuzzyUC struct {
	calls int
}

func (s *countingFuzzyUC) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	s.calls++
	return &dto.SearchResponse{Page: 1, PageSize: 20}, nil
}

// TestMarketplaceAliasIsPublic pins the public alias: GET /api/v1/marketplace
// must list the marketplace through the search use case with no Authorization
// header, so the catalogue is browsable by anyone.
func TestMarketplaceAliasIsPublic(t *testing.T) {
	t.Parallel()

	fuzzy := &countingFuzzyUC{}
	router := NewRouter(
		nil, nil, nil, nil, nil, nil, nil,
		middleware.NewAuthMiddleware(ownershipJWT{}),
		middleware.NewSuspensionMiddleware(stubUserRepo{}),
		nil, handler.NewSearchHandler(fuzzy), nil, nil, nil, nil,
		ws.NewHandler(ws.NewHub(), nil, &stubConversationUC{}),
		nil, nil, nil, nil, nil, nil, nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketplace", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 without auth; body = %s", rr.Code, rr.Body.String())
	}
	if fuzzy.calls != 1 {
		t.Errorf("search calls = %d, want 1: the alias must route to the search listing", fuzzy.calls)
	}
}
