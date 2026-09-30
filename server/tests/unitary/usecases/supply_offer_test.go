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

	baseRequest := supplyTestRequest(testOtherID)
	closedRequest := supplyTestRequest(testOtherID)
	closedRequest.Status = domain.SupplyRequestCancelled
	singleProviderRequest := supplyTestRequest(testOtherID)
	singleProviderRequest.MultipleProviders = false
	lowRemainingRequest := supplyTestRequest(testOtherID)
	lowRemainingRequest.ActualAmount = 10
	duplicateRequest := supplyTestRequest(testOtherID)
	duplicateOffer := supplyTestOffer(testUserID, duplicateRequest.ID)
	ownRequest := supplyTestRequest(testUserID)

	validReq := func(request *domain.SupplyRequest) dto.SupplyOfferDTO {
		return dto.SupplyOfferDTO{
			SupplyRequest:       &request.ID,
			TotalAmount:         20,
			AmountUnit:          domain.Kg,
			PricePerUnit:        ptrFloat64(supplyTestPrice),
			Comments:            "fresh harvest",
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
		// RF-11 lists the price among the fields the offer must include, so these
		// are business rejections: a supplier cannot publish an offer without
		// quoting what it costs. The nullable column is about the rows that
		// predate it, not about this path.
		{
			name:        "missing price",
			ctx:         principalCtx(),
			req:         func() dto.SupplyOfferDTO { r := validReq(&baseRequest); r.PricePerUnit = nil; return r }(),
			seedRequest: &baseRequest,
			wantErr:     domain.ErrInvalidPrice,
		},
		{
			name:        "zero price",
			ctx:         principalCtx(),
			req:         func() dto.SupplyOfferDTO { r := validReq(&baseRequest); r.PricePerUnit = ptrFloat64(0); return r }(),
			seedRequest: &baseRequest,
			wantErr:     domain.ErrInvalidPrice,
		},
		{
			name:        "negative price",
			ctx:         principalCtx(),
			req:         func() dto.SupplyOfferDTO { r := validReq(&baseRequest); r.PricePerUnit = ptrFloat64(-1); return r }(),
			seedRequest: &baseRequest,
			wantErr:     domain.ErrInvalidPrice,
		},
		{name: "request not found", ctx: principalCtx(), req: validReq(&baseRequest), wantErr: domain.ErrNotFound},
		{name: "request not open", ctx: principalCtx(), req: validReq(&closedRequest), seedRequest: &closedRequest, wantErr: domain.ErrInvalidRequestStatus},
		{
			name:        "buyer cannot offer on own request",
			ctx:         principalCtx(),
			req:         validReq(&ownRequest),
			seedRequest: &ownRequest,
			wantErr:     domain.ErrForbidden,
		},
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
				if len(offerRepo.created) != 0 {
					t.Fatalf("repo creates = %d, want 0", len(offerRepo.created))
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
			// The price has to survive the round trip through the constructor and
			// the DTO, in both directions. A price that validates and then is
			// dropped on the way to storage is an offer that a buyer cannot compare.
			if got.PricePerUnit == nil || *got.PricePerUnit != supplyTestPrice {
				t.Errorf("price per unit = %v, want %v", got.PricePerUnit, supplyTestPrice)
			}
			if got.Comments != "fresh harvest" {
				t.Errorf("comments = %q, want %q", got.Comments, "fresh harvest")
			}
			if len(offerRepo.created) != 1 {
				t.Fatalf("saved offers = %d, want 1", len(offerRepo.created))
			}
			if persisted := offerRepo.created[0]; persisted.PricePerUnit == nil || *persisted.PricePerUnit != supplyTestPrice {
				t.Errorf("persisted price per unit = %v, want %v", persisted.PricePerUnit, supplyTestPrice)
			}
			if persisted := offerRepo.created[0]; persisted.Comments != "fresh harvest" {
				t.Errorf("persisted comments = %q, want %q", persisted.Comments, "fresh harvest")
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
		PricePerUnit:        ptrFloat64(7.25),
		Comments:            "bulk price",
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
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 5, PricePerUnit: ptrFloat64(supplyTestPrice)},
			wantErr:     domain.ErrInvalidInput,
		},
		{
			name:        "exceeds remaining amount",
			ctx:         principalCtx(),
			id:          lowRemainingOffer.ID,
			seedRequest: &lowRemainingRequest,
			seedOffer:   &lowRemainingOffer,
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 50, PricePerUnit: ptrFloat64(supplyTestPrice)},
			wantErr:     domain.ErrInsufficientAmount,
		},
		{
			name:        "price dropped on update",
			ctx:         principalCtx(),
			id:          ownedOffer.ID,
			seedRequest: &request,
			seedOffer:   &ownedOffer,
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 30, PricePerUnit: nil},
			wantErr:     domain.ErrInvalidPrice,
		},
		{
			name:        "zero price on update",
			ctx:         principalCtx(),
			id:          ownedOffer.ID,
			seedRequest: &request,
			seedOffer:   &ownedOffer,
			req:         dto.SupplyOfferUpdateDTO{TotalAmount: 30, PricePerUnit: ptrFloat64(0)},
			wantErr:     domain.ErrInvalidPrice,
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
			if saved.PricePerUnit == nil || *saved.PricePerUnit != 7.25 {
				t.Errorf("price per unit = %v, want 7.25", saved.PricePerUnit)
			}
			if saved.Comments != "bulk price" {
				t.Errorf("comments = %q, want %q", saved.Comments, "bulk price")
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

// A passed offer must not come back in the buyer's own listing: Pass rejects it,
// and re-surfacing the same supplier on the same request would let the buyer
// re-pick a candidate they already declined.
func TestSupplyOfferUseCaseListByRequestHidesPassedOffer(t *testing.T) {
	t.Parallel()

	request := supplyTestRequest(testUserID)
	passedOffer := supplyTestOffer(testCompanyID, request.ID)
	keptOffer := supplyTestOffer(testOtherID, request.ID)

	requestRepo := newSupplyFakeRequestRepo()
	requestRepo.requests[request.ID] = request
	offerRepo := newSupplyFakeOfferRepo()
	offerRepo.offers[passedOffer.ID] = passedOffer
	offerRepo.offers[keptOffer.ID] = keptOffer

	// The buyer passes on one supplier.
	matchUC := usecases.NewMatchUseCase(
		requestRepo, offerRepo, newSupplyFakeMatchRepo(), newFakeMatchTransactionRepo(), nil, nil,
	)
	if err := matchUC.Pass(principalCtx(), passedOffer.ID); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if offerRepo.offers[passedOffer.ID].Status != domain.OfferRejected {
		t.Fatalf("passed offer status = %v, want rejected", offerRepo.offers[passedOffer.ID].Status)
	}

	offerUC := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())
	got, err := offerUC.ListByRequest(principalCtx(), request.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, d := range got {
		if d.ID != nil && *d.ID == passedOffer.ID {
			t.Fatalf("passed offer is still listed for the buyer")
		}
	}
	if len(got) != 1 {
		t.Fatalf("listed offers = %d, want 1 (only the offer that was not passed)", len(got))
	}
	if got[0].ID == nil || *got[0].ID != keptOffer.ID {
		t.Errorf("listed offer = %v, want %v", got[0].ID, keptOffer.ID)
	}
}

// The buyer must keep seeing the offer they matched: exclusion applies to
// OfferRejected ONLY, never to OfferMatched.
func TestSupplyOfferUseCaseListByRequestKeepsMatchedOffer(t *testing.T) {
	t.Parallel()

	request := supplyTestRequest(testUserID)
	rejectedOffer := supplyTestOffer(testCompanyID, request.ID)
	matchedOffer := supplyTestOffer(testOtherID, request.ID)
	withdrawnOffer := supplyTestOffer(testUserID, request.ID)
	matchedOffer.Status = domain.OfferMatched
	rejectedOffer.Status = domain.OfferRejected
	withdrawnOffer.Status = domain.OfferWithdrawn

	requestRepo := newSupplyFakeRequestRepo()
	requestRepo.requests[request.ID] = request
	offerRepo := newSupplyFakeOfferRepo()
	offerRepo.offers[rejectedOffer.ID] = rejectedOffer
	offerRepo.offers[matchedOffer.ID] = matchedOffer
	offerRepo.offers[withdrawnOffer.ID] = withdrawnOffer

	uc := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())
	got, err := uc.ListByRequest(principalCtx(), request.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("listed offers = %d, want 2 (matched and withdrawn survive)", len(got))
	}
	statuses := map[uuid.UUID]domain.OfferStatus{}
	for _, d := range got {
		if d.ID == nil {
			t.Fatal("listed offer has nil id")
		}
		statuses[*d.ID] = d.Status
	}
	if _, ok := statuses[rejectedOffer.ID]; ok {
		t.Errorf("rejected offer is still listed")
	}
	if statuses[matchedOffer.ID] != domain.OfferMatched {
		t.Errorf("matched offer status = %v, want matched", statuses[matchedOffer.ID])
	}
	if statuses[withdrawnOffer.ID] != domain.OfferWithdrawn {
		t.Errorf("withdrawn offer status = %v, want withdrawn", statuses[withdrawnOffer.ID])
	}
}

// The supplier keeps full visibility of their own rejected offer: a different
// audience with a different information need. Only the buyer's read hides it.
func TestSupplyOfferUseCaseListBySupplierStillShowsRejectedOffer(t *testing.T) {
	t.Parallel()

	rejected := supplyTestOffer(testUserID, uuid.New())
	rejected.Status = domain.OfferRejected

	offerRepo := newSupplyFakeOfferRepo()
	offerRepo.offers[rejected.ID] = rejected

	uc := usecases.NewSupplyOfferUseCase(offerRepo, newSupplyFakeRequestRepo(), newSupplyFakeMatchRepo(), newFakeTimer())
	got, err := uc.ListBySupplier(principalCtx(), testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("listed offers = %d, want 1", len(got))
	}
	if got[0].Status != domain.OfferRejected {
		t.Errorf("status = %v, want rejected", got[0].Status)
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
