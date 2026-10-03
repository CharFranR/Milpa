package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
)

type stubMatchUC struct {
	err          error
	match        *dto.MatchDTO
	transaction  *dto.TransactionDTO
	list         []*dto.MatchDTO
	prioritized  []*dto.PrioritizedOfferDTO
	gotOfferID   uuid.UUID
	gotMatchID   uuid.UUID
	gotRequestID uuid.UUID
}

func (s *stubMatchUC) Like(ctx context.Context, supplyOfferID uuid.UUID) (*dto.MatchDTO, *dto.TransactionDTO, error) {
	s.gotOfferID = supplyOfferID
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.match, s.transaction, nil
}

func (s *stubMatchUC) Pass(ctx context.Context, supplyOfferID uuid.UUID) error {
	s.gotOfferID = supplyOfferID
	return s.err
}

func (s *stubMatchUC) GetByID(ctx context.Context, matchID uuid.UUID) (*dto.MatchDTO, error) {
	s.gotMatchID = matchID
	if s.err != nil {
		return nil, s.err
	}
	return s.match, nil
}

func (s *stubMatchUC) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.MatchDTO, error) {
	s.gotRequestID = supplyRequestID
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubMatchUC) ListPrioritized(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error) {
	s.gotRequestID = supplyRequestID
	if s.err != nil {
		return nil, s.err
	}
	return s.prioritized, nil
}

type stubRecommendationUC struct {
	err         error
	quantity    float64
	gotSupplier uuid.UUID
	gotProduct  string
}

func (s *stubRecommendationUC) RankOffers(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error) {
	return nil, s.err
}

func (s *stubRecommendationUC) AvailableQuantity(ctx context.Context, supplierID uuid.UUID, productName string) (float64, error) {
	s.gotSupplier = supplierID
	s.gotProduct = productName
	return s.quantity, s.err
}

func newMatchTestRouter(t *testing.T, matchUC primary.MatchUseCase, recUC primary.RecommendationUseCase) http.Handler {
	t.Helper()

	chat := ws.NewHandler(ws.NewHub(), nil, &stubConversationUC{})
	authMW := middleware.NewAuthMiddleware(stubJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	return NewRouter(nil, nil, nil, nil, nil, nil, nil, authMW, suspensionMW, nil, nil, nil, nil, nil, nil, chat, nil, nil, nil, handler.NewMatchHandler(matchUC), handler.NewRecommendationHandler(recUC))
}

func TestMatchRoutesRequireToken(t *testing.T) {
	t.Parallel()

	offerID := uuid.NewString()
	requestID := uuid.NewString()
	matchID := uuid.NewString()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/matches/like/" + offerID},
		{http.MethodPost, "/api/v1/matches/pass/" + offerID},
		{http.MethodGet, "/api/v1/matches/requests/" + requestID},
		{http.MethodGet, "/api/v1/matches/requests/" + requestID + "/prioritized"},
		{http.MethodGet, "/api/v1/matches/" + matchID},
		{http.MethodGet, "/api/v1/recommendations/availability?supplier_id=" + uuid.NewString() + "&product_name=Maize"},
	}

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{})

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (route missing or auth not attached); body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestLikeRouteCreatesMatchAndExtractsOfferID(t *testing.T) {
	t.Parallel()

	offerID := uuid.New()
	matchID := uuid.New()
	transactionID := uuid.New()
	matchUC := &stubMatchUC{
		match:       &dto.MatchDTO{ID: matchID, SupplyOffer: offerID, MatchedAmount: 30},
		transaction: &dto.TransactionDTO{ID: &transactionID, MatchID: &matchID},
	}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/like/"+offerID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotOfferID != offerID {
		t.Fatalf("use case received offer id %v, want %v — route param name does not match chi.URLParam", matchUC.gotOfferID, offerID)
	}

	var body struct {
		Data struct {
			Match       dto.MatchDTO       `json:"match"`
			Transaction dto.TransactionDTO `json:"transaction"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.Match.ID != matchID || body.Data.Transaction.ID == nil || *body.Data.Transaction.ID != transactionID {
		t.Fatalf("body = %+v, want the match and transaction returned by the use case", body.Data)
	}
}

func TestLikeRouteStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		ucErr  error
		status int
	}{
		{"conflict when offer is not actionable", domain.ErrInvalidOfferStatus, http.StatusConflict},
		{"conflict when request is not open", domain.ErrInvalidRequestStatus, http.StatusConflict},
		{"conflict when a match already exists", domain.ErrInvalidMatchStatus, http.StatusConflict},
		{"conflict when availability is insufficient", domain.ErrInsufficientAmount, http.StatusConflict},
		{"forbidden for non-buyer", domain.ErrForbidden, http.StatusForbidden},
		{"not found for missing offer", domain.ErrNotFound, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newMatchTestRouter(t, &stubMatchUC{err: tt.ucErr}, &stubRecommendationUC{})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/like/"+uuid.NewString(), nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.status, rr.Body.String())
			}
		})
	}
}

func TestLikeRouteRejectsInvalidOfferID(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/like/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
	}
}

func TestPassRouteRejectsOffer(t *testing.T) {
	t.Parallel()

	offerID := uuid.New()
	matchUC := &stubMatchUC{}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/pass/"+offerID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotOfferID != offerID {
		t.Fatalf("use case received offer id %v, want %v", matchUC.gotOfferID, offerID)
	}
}

func TestPassRouteConflictsWhenOfferNotActionable(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{err: domain.ErrInvalidOfferStatus}, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/pass/"+uuid.NewString(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rr.Code, rr.Body.String())
	}
}

func TestListMatchesByRequestRoute(t *testing.T) {
	t.Parallel()

	requestID := uuid.New()
	matchID := uuid.New()
	matchUC := &stubMatchUC{list: []*dto.MatchDTO{{ID: matchID, MatchedAmount: 30}}}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/requests/"+requestID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotRequestID != requestID {
		t.Fatalf("use case received request id %v, want %v", matchUC.gotRequestID, requestID)
	}
}

func TestPrioritizedRouteReturnsRankedOffers(t *testing.T) {
	t.Parallel()

	requestID := uuid.New()
	matchUC := &stubMatchUC{
		prioritized: []*dto.PrioritizedOfferDTO{{Score: 42, AvailableQuantity: 20}},
	}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/requests/"+requestID.String()+"/prioritized", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotRequestID != requestID {
		t.Fatalf("use case received request id %v, want %v", matchUC.gotRequestID, requestID)
	}

	var body struct {
		Data []struct {
			Score float64 `json:"score"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].Score != 42 {
		t.Fatalf("body = %+v, want the prioritized list", body.Data)
	}
}

func TestGetMatchRouteExtractsMatchID(t *testing.T) {
	t.Parallel()

	matchID := uuid.New()
	matchUC := &stubMatchUC{match: &dto.MatchDTO{ID: matchID}}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+matchID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotMatchID != matchID {
		t.Fatalf("use case received match id %v, want %v — route param name does not match chi.URLParam", matchUC.gotMatchID, matchID)
	}
}

func TestGetMatchRouteForbiddenForThirdParty(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{err: domain.ErrForbidden}, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+uuid.NewString(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", rr.Code, rr.Body.String())
	}
}

func TestAvailabilityRoute(t *testing.T) {
	t.Parallel()

	supplierID := uuid.New()
	recUC := &stubRecommendationUC{quantity: 20}
	router := newMatchTestRouter(t, &stubMatchUC{}, recUC)

	query := url.Values{"supplier_id": {supplierID.String()}, "product_name": {"Maize"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/availability?"+query.Encode(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}
	if recUC.gotSupplier != supplierID || recUC.gotProduct != "Maize" {
		t.Fatalf("use case received %v/%q, want %v/Maize", recUC.gotSupplier, recUC.gotProduct, supplierID)
	}

	var body struct {
		Data struct {
			AvailableQuantity float64 `json:"available_quantity"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.AvailableQuantity != 20 {
		t.Fatalf("available quantity = %v, want 20", body.Data.AvailableQuantity)
	}
}

func TestAvailabilityRouteValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		query  string
		status int
	}{
		{"missing supplier id", "product_name=Maize", http.StatusBadRequest},
		{"invalid supplier id", "supplier_id=nope&product_name=Maize", http.StatusBadRequest},
		{"missing product name", "supplier_id=" + uuid.NewString(), http.StatusBadRequest},
	}

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/availability?"+tt.query, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, tt.status, rr.Body.String())
			}
		})
	}
}

func TestAvailabilityRouteNotFound(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{err: domain.ErrNotFound})

	query := url.Values{"supplier_id": {uuid.NewString()}, "product_name": {"Maize"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/availability?"+query.Encode(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rr.Code, rr.Body.String())
	}
}

func TestRecommendationGroupUnknownPathIs404(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/unknown", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rr.Code, rr.Body.String())
	}
}

func TestMatchesGroupStaticRequestPathWinsOverMatchID(t *testing.T) {
	t.Parallel()

	matchUC := &stubMatchUC{}
	router := newMatchTestRouter(t, matchUC, &stubRecommendationUC{})

	requestID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/requests/"+requestID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (static 'requests' segment must not be captured by {matchID}); body = %s", rr.Code, rr.Body.String())
	}
	if matchUC.gotRequestID != requestID {
		t.Fatalf("use case received request id %v, want %v", matchUC.gotRequestID, requestID)
	}
}

func TestMatchRouteInvalidUUIDReturns400(t *testing.T) {
	t.Parallel()

	router := newMatchTestRouter(t, &stubMatchUC{}, &stubRecommendationUC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rr.Code, rr.Body.String())
	}
}
