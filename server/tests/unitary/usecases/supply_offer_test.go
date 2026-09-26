package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	"milpa/internal/auth"
)

func TestSupplyOfferUseCaseCreate(t *testing.T) {
	t.Parallel()

	baseRequest := supplyTestRequest(testUserID)
	closedRequest := supplyTestRequest(testUserID)
	closedRequest.Status = domain.SupplyRequestCancelled
	singleProviderRequest := supplyTestRequest(testUserID)
	singleProviderRequest.MultipleProviders = false
	lowRemainingRequest := supplyTestRequest(testUserID)
	lowRemainingRequest.ActualAmount = 10
	duplicateRequest := supplyTestRequest(testUserID)
	duplicateOffer := supplyTestOffer(testUserID, duplicateRequest.ID)

	validReq := func(request *domain.SupplyRequest) dto.SupplyOfferDTO {
		return dto.SupplyOfferDTO{
			SupplyRequest:       &request.ID,
			TotalAmount:         20,
			AmountUnit:          domain.Kg,
			ProposedDeliveryDay: fixedTime.Add(48 * time.Hour),
			DeliveryAvailable:   true,
		}
	}

	tests := []struct {
		name        string
		ctx         context.Context
		req         dto.SupplyOfferDTO
		seedRequest *domain.SupplyRequest
		seedOffer   *domain.SupplyOffer
		findErr     error
		matchActive bool
		createErr   error
		wantErr     error
	}{
		{name: "happy path", ctx: principalCtx(), req: validReq(&baseRequest), seedRequest: &baseRequest},
		{name: "unauthenticated", ctx: context.Background(), req: validReq(&baseRequest), seedRequest: &baseRequest, wantErr: auth.ErrUnauthenticated},
		{name: "missing request id", ctx: principalCtx(), req: dto.SupplyOfferDTO{TotalAmount: 20}, wantErr: domain.ErrInvalidInput},
		{name: "zero amount", ctx: principalCtx(), req: dto.SupplyOfferDTO{SupplyRequest: &baseRequest.ID}, wantErr: domain.ErrInvalidInput},
		{name: "request not found", ctx: principalCtx(), req: validReq(&baseRequest), wantErr: domain.ErrNotFound},
		{name: "request not open", ctx: principalCtx(), req: validReq(&closedRequest), seedRequest: &closedRequest, wantErr: domain.ErrInvalidRequestStatus},
		{
			name:        "one offer per supplier per request",
			ctx:         principalCtx(),
			req:         validReq(&duplicateRequest),
			seedRequest: &duplicateRequest,
			seedOffer:   &duplicateOffer,
			wantErr:     domain.ErrDuplicate,
		},
		{name: "find repo error", ctx: principalCtx(), req: validReq(&baseRequest), seedRequest: &baseRequest, findErr: errFake, wantErr: errFake},
		{
			name:        "single provider with active match",
			ctx:         principalCtx(),
			req:         validReq(&singleProviderRequest),
			seedRequest: &singleProviderRequest,
			matchActive: true,
			wantErr:     primary.ErrActiveMatch,
		},
		{
			name:        "single provider without active match",
			ctx:         principalCtx(),
			req:         validReq(&singleProviderRequest),
			seedRequest: &singleProviderRequest,
		},
		{
			name:        "below minimum per provider",
			ctx:         principalCtx(),
			req:         func() dto.SupplyOfferDTO { r := validReq(&baseRequest); r.TotalAmount = 5; return r }(),
			seedRequest: &baseRequest,
			wantErr:     domain.ErrInvalidInput,
		},
		{
			name:        "exceeds remaining amount",
			ctx:         principalCtx(),
			req:         validReq(&lowRemainingRequest),
			seedRequest: &lowRemainingRequest,
			wantErr:     domain.ErrInsufficientAmount,
		},
		{
			name:        "repo duplicate on create",
			ctx:         principalCtx(),
			req:         validReq(&baseRequest),
			seedRequest: &baseRequest,
			createErr:   domain.ErrDuplicate,
			wantErr:     domain.ErrDuplicate,
		},
		{name: "repo error", ctx: principalCtx(), req: validReq(&baseRequest), seedRequest: &baseRequest, createErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seedRequest != nil {
				requestRepo.requests[tt.seedRequest.ID] = *tt.seedRequest
			}
			offerRepo := newSupplyFakeOfferRepo()
			if tt.seedOffer != nil {
				offerRepo.offers[tt.seedOffer.ID] = *tt.seedOffer
			}
			offerRepo.findErr = tt.findErr
			offerRepo.createErr = tt.createErr
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.existsActive = tt.matchActive
			uc := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, matchRepo, newFakeTimer())

			got, err := uc.Create(tt.ctx, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == nil || *got.ID == uuid.Nil {
				t.Fatal("expected a generated ID, got nil UUID")
			}
			if got.SupplierID == nil || *got.SupplierID != testUserID {
				t.Errorf("supplier id = %v, want principal %v", got.SupplierID, testUserID)
			}
			if got.Status != domain.OfferActive {
				t.Errorf("status = %v, want active", got.Status)
			}
			if len(offerRepo.created) != 1 {
				t.Fatalf("saved offers = %d, want 1", len(offerRepo.created))
			}
		})
	}
}

func TestSupplyOfferUseCaseUpdate(t *testing.T) {
	t.Parallel()

	request := supplyTestRequest(testUserID)
	ownedOffer := supplyTestOffer(testUserID, request.ID)
	foreignOffer := supplyTestOffer(testOtherID, request.ID)
	matchedOffer := supplyTestOffer(testUserID, request.ID)
	matchedOffer.Status = domain.OfferMatched

	lowMinRequest := supplyTestRequest(testUserID)
	lowMinOffer := supplyTestOffer(testUserID, lowMinRequest.ID)
	lowRemainingRequest := supplyTestRequest(testUserID)
	lowRemainingRequest.ActualAmount = 10
	lowRemainingOffer := supplyTestOffer(testUserID, lowRemainingRequest.ID)

	validReq := dto.SupplyOfferUpdateDTO{
		TotalAmount:         30,
		AmountUnit:          domain.Lb,
		ProposedDeliveryDay: fixedTime.Add(50 * time.Hour),
		DeliveryAvailable:   false,
	}

	tests := []struct {
		name        string
		ctx         context.Context
		id          uuid.UUID
		seedRequest *domain.SupplyRequest
		seedOffer   *domain.SupplyOffer
		req         dto.SupplyOfferUpdateDTO
		updateErr   error
		wantErr     error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedOffer.ID, seedRequest: &request, seedOffer: &ownedOffer, req: validReq},
		{name: "unauthenticated", ctx: context.Background(), id: ownedOffer.ID, seedRequest: &request, seedOffer: &ownedOffer, req: validReq, wantErr: auth.ErrUnauthenticated},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, req: validReq, wantErr: domain.ErrInvalidInput},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), seedRequest: &request, req: validReq, wantErr: domain.ErrNotFound},
		{name: "non-owner", ctx: principalCtx(), id: foreignOffer.ID, seedRequest: &request, seedOffer: &foreignOffer, req: validReq, wantErr: domain.ErrForbidden},
		{name: "not actionable", ctx: principalCtx(), id: matchedOffer.ID, seedRequest: &request, seedOffer: &matchedOffer, req: validReq, wantErr: domain.ErrInvalidOfferStatus},
		{name: "zero amount", ctx: principalCtx(), id: ownedOffer.ID, seedRequest: &request, seedOffer: &ownedOffer, req: dto.SupplyOfferUpdateDTO{}, wantErr: domain.ErrInvalidInput},
		{name: "request missing for policy check", ctx: principalCtx(), id: ownedOffer.ID, seedOffer: &ownedOffer, req: validReq, wantErr: domain.ErrNotFound},
		{
			name:        "below minimum per provider",
			ctx:         principalCtx(),
			id:          lowMinOffer.ID,
			seedRequest: &lowMinRequest,
			seedOffer:   &lowMinOffer,
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 5},
			wantErr:     domain.ErrInvalidInput,
		},
		{
			name:        "exceeds remaining amount",
			ctx:         principalCtx(),
			id:          lowRemainingOffer.ID,
			seedRequest: &lowRemainingRequest,
			seedOffer:   &lowRemainingOffer,
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 50},
			wantErr:     domain.ErrInsufficientAmount,
		},
		{name: "repo error", ctx: principalCtx(), id: ownedOffer.ID, seedRequest: &request, seedOffer: &ownedOffer, req: validReq, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seedRequest != nil {
				requestRepo.requests[tt.seedRequest.ID] = *tt.seedRequest
			}
			offerRepo := newSupplyFakeOfferRepo()
			if tt.seedOffer != nil {
				offerRepo.offers[tt.seedOffer.ID] = *tt.seedOffer
			}
			offerRepo.updateErr = tt.updateErr
			uc := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

			err := uc.Update(tt.ctx, tt.id, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(offerRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(offerRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := offerRepo.offers[tt.id]
			if saved.TotalAmount != 30 || saved.AmountUnit != domain.Lb {
				t.Errorf("amounts = %v / %v, want 30 / Lb", saved.TotalAmount, saved.AmountUnit)
			}
			if saved.DeliveryAvailable {
				t.Error("delivery available = true, want false")
			}
			if !saved.ProposedDeliveryDay.Equal(validReq.ProposedDeliveryDay) {
				t.Errorf("proposed delivery day = %v, want %v", saved.ProposedDeliveryDay, validReq.ProposedDeliveryDay)
			}
			if !saved.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", saved.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestSupplyOfferUseCaseWithdraw(t *testing.T) {
	t.Parallel()

	request := supplyTestRequest(testUserID)
	ownedOffer := supplyTestOffer(testUserID, request.ID)
	foreignOffer := supplyTestOffer(testOtherID, request.ID)
	matchedOffer := supplyTestOffer(testUserID, request.ID)
	matchedOffer.Status = domain.OfferMatched

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		seedOffer *domain.SupplyOffer
		updateErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedOffer.ID, seedOffer: &ownedOffer},
		{name: "unauthenticated", ctx: context.Background(), id: ownedOffer.ID, seedOffer: &ownedOffer, wantErr: auth.ErrUnauthenticated},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "non-owner", ctx: principalCtx(), id: foreignOffer.ID, seedOffer: &foreignOffer, wantErr: domain.ErrForbidden},
		{name: "matched offer cannot withdraw", ctx: principalCtx(), id: matchedOffer.ID, seedOffer: &matchedOffer, wantErr: domain.ErrInvalidOfferStatus},
		{name: "repo error", ctx: principalCtx(), id: ownedOffer.ID, seedOffer: &ownedOffer, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offerRepo := newSupplyFakeOfferRepo()
			if tt.seedOffer != nil {
				offerRepo.offers[tt.seedOffer.ID] = *tt.seedOffer
			}
			offerRepo.updateErr = tt.updateErr
			uc := usecases.NewSupplyOfferUseCase(offerRepo, newSupplyFakeRequestRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			err := uc.Withdraw(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(offerRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(offerRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := offerRepo.offers[tt.id]
			if saved.Status != domain.OfferWithdrawn {
				t.Errorf("status = %v, want withdrawn", saved.Status)
			}
		})
	}
}

func TestSupplyOfferUseCaseGetByID(t *testing.T) {
	t.Parallel()

	ownedOffer := supplyTestOffer(testUserID, uuid.New())
	buyerVisibleOffer := supplyTestOffer(testCompanyID, uuid.New())
	buyerRequest := supplyTestRequest(testUserID)
	buyerVisibleOffer.SupplyRequest = buyerRequest.ID
	foreignOffer := supplyTestOffer(testCompanyID, uuid.New())
	strangerRequest := supplyTestRequest(testOtherID)
	foreignOffer.SupplyRequest = strangerRequest.ID

	tests := []struct {
		name        string
		ctx         context.Context
		id          uuid.UUID
		seedRequest *domain.SupplyRequest
		seedOffer   *domain.SupplyOffer
		wantErr     error
	}{
		{name: "supplier reads own offer", ctx: principalCtx(), id: ownedOffer.ID, seedOffer: &ownedOffer},
		{name: "request buyer reads offer", ctx: principalCtx(), id: buyerVisibleOffer.ID, seedRequest: &buyerRequest, seedOffer: &buyerVisibleOffer},
		{name: "stranger forbidden", ctx: principalCtx(), id: foreignOffer.ID, seedRequest: &strangerRequest, seedOffer: &foreignOffer, wantErr: domain.ErrForbidden},
		{name: "request missing for stranger check", ctx: principalCtx(), id: foreignOffer.ID, seedOffer: &foreignOffer, wantErr: domain.ErrNotFound},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, wantErr: domain.ErrInvalidInput},
		{name: "unauthenticated", ctx: context.Background(), id: ownedOffer.ID, wantErr: auth.ErrUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seedRequest != nil {
				requestRepo.requests[tt.seedRequest.ID] = *tt.seedRequest
			}
			offerRepo := newSupplyFakeOfferRepo()
			if tt.seedOffer != nil {
				offerRepo.offers[tt.seedOffer.ID] = *tt.seedOffer
			}
			uc := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.GetByID(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == nil || *got.ID != tt.id {
				t.Errorf("id = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestSupplyOfferUseCaseListByRequest(t *testing.T) {
	t.Parallel()

	request := supplyTestRequest(testUserID)
	strangerRequest := supplyTestRequest(testOtherID)
	ownOffer := supplyTestOffer(testUserID, request.ID)
	otherSupplierOffer := supplyTestOffer(testCompanyID, request.ID)
	unrelatedOffer := supplyTestOffer(testCompanyID, strangerRequest.ID)

	tests := []struct {
		name        string
		ctx         context.Context
		requestID   uuid.UUID
		seedRequest *domain.SupplyRequest
		seedOffers  []domain.SupplyOffer
		wantLen     int
		wantErr     error
	}{
		{
			name:        "buyer lists offers on own request",
			ctx:         principalCtx(),
			requestID:   request.ID,
			seedRequest: &request,
			seedOffers:  []domain.SupplyOffer{ownOffer, otherSupplierOffer, unrelatedOffer},
			wantLen:     2,
		},
		{name: "non-owner forbidden", ctx: principalCtx(), requestID: strangerRequest.ID, seedRequest: &strangerRequest, wantErr: domain.ErrForbidden},
		{name: "request not found", ctx: principalCtx(), requestID: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "null id", ctx: principalCtx(), requestID: uuid.Nil, wantErr: domain.ErrInvalidInput},
		{name: "unauthenticated", ctx: context.Background(), requestID: request.ID, wantErr: auth.ErrUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seedRequest != nil {
				requestRepo.requests[tt.seedRequest.ID] = *tt.seedRequest
			}
			offerRepo := newSupplyFakeOfferRepo()
			for _, supplyOffer := range tt.seedOffers {
				offerRepo.offers[supplyOffer.ID] = supplyOffer
			}
			uc := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.ListByRequest(tt.ctx, tt.requestID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("listed offers = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSupplyOfferUseCaseListBySupplier(t *testing.T) {
	t.Parallel()

	ownOffer := supplyTestOffer(testUserID, uuid.New())
	secondOwnOffer := supplyTestOffer(testUserID, uuid.New())
	foreignOffer := supplyTestOffer(testOtherID, uuid.New())

	tests := []struct {
		name       string
		ctx        context.Context
		supplierID uuid.UUID
		seedOffers []domain.SupplyOffer
		wantLen    int
		wantErr    error
	}{
		{
			name:       "supplier lists own offers",
			ctx:        principalCtx(),
			supplierID: testUserID,
			seedOffers: []domain.SupplyOffer{ownOffer, secondOwnOffer, foreignOffer},
			wantLen:    2,
		},
		{name: "cannot list another supplier", ctx: principalCtx(), supplierID: testOtherID, wantErr: domain.ErrForbidden},
		{name: "unauthenticated", ctx: context.Background(), supplierID: testUserID, wantErr: auth.ErrUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offerRepo := newSupplyFakeOfferRepo()
			for _, supplyOffer := range tt.seedOffers {
				offerRepo.offers[supplyOffer.ID] = supplyOffer
			}
			uc := usecases.NewSupplyOfferUseCase(offerRepo, newSupplyFakeRequestRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.ListBySupplier(tt.ctx, tt.supplierID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("listed offers = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
