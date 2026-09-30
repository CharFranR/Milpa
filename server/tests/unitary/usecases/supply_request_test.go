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

func TestSupplyRequestUseCaseCreate(t *testing.T) {
	t.Parallel()

	validReq := dto.SupplyRequestDTO{
		ProductName:          "Rice",
		TotalAmount:          100,
		AmountUnit:           domain.Kg,
		NumberOfUnits:        10,
		AmountPerUnit:        10,
		UnitOfMeasure:        domain.Kg,
		Address:              domain.Address{Department: "Masaya", Municipality: "Masaya", AddressLine: "Km 5 Carretera Sur"},
		RequestDeadline:      fixedTime.Add(24 * time.Hour),
		DeliveryDeadline:     fixedTime.Add(72 * time.Hour),
		Description:          "Fresh harvest",
		MultipleProviders:    true,
		MinAmountPerProvider: 10,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		req       dto.SupplyRequestDTO
		createErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), req: validReq},
		{name: "unauthenticated", ctx: context.Background(), req: validReq, wantErr: auth.ErrUnauthenticated},
		{name: "empty product name", ctx: principalCtx(), req: dto.SupplyRequestDTO{TotalAmount: 100}, wantErr: domain.ErrInvalidInput},
		{name: "zero total amount", ctx: principalCtx(), req: dto.SupplyRequestDTO{ProductName: "Rice"}, wantErr: domain.ErrInvalidInput},
		{
			name: "request deadline after delivery deadline",
			ctx:  principalCtx(),
			req: dto.SupplyRequestDTO{
				ProductName:      "Rice",
				TotalAmount:      100,
				RequestDeadline:  fixedTime.Add(72 * time.Hour),
				DeliveryDeadline: fixedTime.Add(24 * time.Hour),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{name: "repo error", ctx: principalCtx(), req: validReq, createErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			requestRepo.createErr = tt.createErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

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
			if got.BuyerID == nil || *got.BuyerID != testUserID {
				t.Errorf("buyer id = %v, want principal %v", got.BuyerID, testUserID)
			}
			if got.Status != domain.SupplyRequestOpen {
				t.Errorf("status = %v, want open", got.Status)
			}
			if got.ActualAmount != 100 {
				t.Errorf("actual amount = %v, want 100", got.ActualAmount)
			}
			if got.MinAmountPerProvider != 10 {
				t.Errorf("min amount per provider = %v, want 10", got.MinAmountPerProvider)
			}
			if len(requestRepo.created) != 1 {
				t.Fatalf("saved supply requests = %d, want 1", len(requestRepo.created))
			}
			if requestRepo.created[0].BuyerID != testUserID {
				t.Errorf("saved buyer id = %v, want %v", requestRepo.created[0].BuyerID, testUserID)
			}
		})
	}
}

// TestSupplyRequestUseCaseCreateRequiresUnitAndLocation covers the two required
// fields RF-10 names that Create did not check: the unit of measure and the
// location.
//
// UnitOfMeasure is a 3-value iota decoded straight off the wire, so a client can
// request unit 7 and the repository would store it. The address was passed
// through untouched, and an address with no data is written as
// address_id = NULL, so a request that declared no location at all was accepted
// as a valid, publishable request.
func TestSupplyRequestUseCaseCreateRequiresUnitAndLocation(t *testing.T) {
	t.Parallel()

	valid := dto.SupplyRequestDTO{
		ProductName:      "Rice",
		TotalAmount:      100,
		AmountUnit:       domain.Kg,
		NumberOfUnits:    10,
		AmountPerUnit:    10,
		UnitOfMeasure:    domain.Kg,
		Address:          domain.Address{Department: "Masaya", Municipality: "Masaya", AddressLine: "Km 5 Carretera Sur"},
		RequestDeadline:  fixedTime.Add(24 * time.Hour),
		DeliveryDeadline: fixedTime.Add(72 * time.Hour),
	}

	tests := []struct {
		name    string
		mutate  func(req *dto.SupplyRequestDTO)
		wantErr error
	}{
		{
			name:    "unit of measure outside the vocabulary",
			mutate:  func(req *dto.SupplyRequestDTO) { req.UnitOfMeasure = domain.MeasurementOptions(7) },
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "negative unit of measure",
			mutate:  func(req *dto.SupplyRequestDTO) { req.UnitOfMeasure = domain.MeasurementOptions(-1) },
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "entirely empty address",
			mutate:  func(req *dto.SupplyRequestDTO) { req.Address = domain.Address{} },
			wantErr: domain.ErrDepartmentRequired,
		},
		{
			name:    "address with a line but no department",
			mutate:  func(req *dto.SupplyRequestDTO) { req.Address = domain.Address{AddressLine: "Km 5 Carretera Sur"} },
			wantErr: domain.ErrDepartmentRequired,
		},
		{
			name:   "every required field present",
			mutate: func(req *dto.SupplyRequestDTO) {},
		},
		{
			name:   "every unit in the vocabulary is accepted",
			mutate: func(req *dto.SupplyRequestDTO) { req.UnitOfMeasure = domain.Tn; req.AmountUnit = domain.Tn },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := valid
			tt.mutate(&req)

			requestRepo := newSupplyFakeRequestRepo()
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.Create(principalCtx(), req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				// The repository is never reached: nothing is persisted for a
				// request the brief does not consider well formed.
				if len(requestRepo.created) != 0 {
					t.Errorf("created supply requests = %d, want 0", len(requestRepo.created))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.ID == nil || *got.ID == uuid.Nil {
				t.Fatal("expected a persisted request with a generated ID")
			}
			if len(requestRepo.created) != 1 {
				t.Fatalf("created supply requests = %d, want 1", len(requestRepo.created))
			}
		})
	}
}

func TestValidMeasurementOptions(t *testing.T) {
	t.Parallel()

	for _, m := range []domain.MeasurementOptions{domain.Kg, domain.Lb, domain.Tn} {
		if !domain.ValidMeasurementOptions(m) {
			t.Errorf("ValidMeasurementOptions(%d) = false, want true", int(m))
		}
	}
	for _, m := range []domain.MeasurementOptions{-1, 3, 4, 99} {
		if domain.ValidMeasurementOptions(m) {
			t.Errorf("ValidMeasurementOptions(%d) = true, want false", int(m))
		}
	}
}

func TestSupplyRequestUseCaseList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		seed    []domain.SupplyRequest
		listErr error
		wantLen int
		wantErr error
	}{
		{
			name:    "happy path",
			ctx:     principalCtx(),
			seed:    []domain.SupplyRequest{supplyTestRequest(testUserID), supplyTestRequest(testUserID), supplyTestRequest(testOtherID)},
			wantLen: 2,
		},
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "repo error", ctx: principalCtx(), listErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			for _, supplyRequest := range tt.seed {
				requestRepo.requests[supplyRequest.ID] = supplyRequest
			}
			requestRepo.listErr = tt.listErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.List(tt.ctx)

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
				t.Fatalf("listed supply requests = %d, want %d", len(got), tt.wantLen)
			}
			if requestRepo.listBuyer != testUserID {
				t.Errorf("list buyer id = %v, want principal %v", requestRepo.listBuyer, testUserID)
			}
		})
	}
}

func TestSupplyRequestUseCaseListAvailable(t *testing.T) {
	t.Parallel()

	ownRequest := supplyTestRequest(testUserID)
	otherBuyerRequest := supplyTestRequest(testOtherID)
	zeroAmountRequest := supplyTestRequest(testOtherID)
	zeroAmountRequest.ActualAmount = 0
	offeredRequest := supplyTestRequest(testOtherID)
	competitorOfferedRequest := supplyTestRequest(testOtherID)

	tests := []struct {
		name        string
		ctx         context.Context
		supplierID  uuid.UUID
		seed        []domain.SupplyRequest
		seedOffers  []domain.SupplyOffer
		listOpenErr error
		listErr     error
		wantLen     int
		wantErr     error
	}{
		{
			name:       "returns other buyers open requests",
			ctx:        principalCtx(),
			supplierID: testUserID,
			seed:       []domain.SupplyRequest{ownRequest, otherBuyerRequest, competitorOfferedRequest},
			seedOffers: []domain.SupplyOffer{supplyTestOffer(testCompanyID, competitorOfferedRequest.ID)},
			wantLen:    2,
		},
		{name: "own requests excluded", ctx: principalCtx(), supplierID: testUserID, seed: []domain.SupplyRequest{ownRequest}, wantLen: 0},
		{name: "zero actual amount excluded", ctx: principalCtx(), supplierID: testUserID, seed: []domain.SupplyRequest{zeroAmountRequest}, wantLen: 0},
		{
			name:       "requests already offered by supplier excluded",
			ctx:        principalCtx(),
			supplierID: testUserID,
			seed:       []domain.SupplyRequest{offeredRequest},
			seedOffers: []domain.SupplyOffer{supplyTestOffer(testUserID, offeredRequest.ID)},
			wantLen:    0,
		},
		{name: "supplier mismatch", ctx: principalCtx(), supplierID: testOtherID, wantErr: domain.ErrForbidden},
		{name: "unauthenticated", ctx: context.Background(), supplierID: testUserID, wantErr: auth.ErrUnauthenticated},
		{name: "list open repo error", ctx: principalCtx(), supplierID: testUserID, listOpenErr: errFake, wantErr: errFake},
		{name: "offer list repo error", ctx: principalCtx(), supplierID: testUserID, listErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			for _, supplyRequest := range tt.seed {
				requestRepo.requests[supplyRequest.ID] = supplyRequest
			}
			requestRepo.listOpenErr = tt.listOpenErr
			offerRepo := newSupplyFakeOfferRepo()
			for _, supplyOffer := range tt.seedOffers {
				offerRepo.offers[supplyOffer.ID] = supplyOffer
			}
			offerRepo.listErr = tt.listErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, offerRepo, newSupplyFakeMatchRepo(), newFakeTimer())

			got, err := uc.ListAvailable(tt.ctx, tt.supplierID)

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
			if got == nil {
				t.Fatal("result slice is nil, want non-nil slice")
			}
			if len(got) != tt.wantLen {
				t.Fatalf("available supply requests = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSupplyRequestUseCaseGetByID(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)

	tests := []struct {
		name    string
		ctx     context.Context
		id      uuid.UUID
		seed    *domain.SupplyRequest
		getErr  error
		wantErr error
	}{
		{name: "owner reads own request", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest},
		{name: "any authenticated user can read", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, wantErr: domain.ErrInvalidInput},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, wantErr: auth.ErrUnauthenticated},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, getErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.getErr = tt.getErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

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

func TestSupplyRequestUseCaseUpdate(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)
	cancelledRequest := supplyTestRequest(testUserID)
	cancelledRequest.Status = domain.SupplyRequestCancelled

	validReq := dto.SupplyGeneralUpdateDTO{
		ProductName:          "Maize",
		TotalAmount:          80,
		ActualAmount:         60,
		AmountUnit:           domain.Lb,
		NumberOfUnits:        8,
		AmountPerUnit:        10,
		UnitOfMeasure:        domain.Lb,
		Address:              domain.Address{Department: "León", Municipality: "León", AddressLine: "Central Market"},
		RequestDeadline:      fixedTime.Add(30 * time.Hour),
		DeliveryDeadline:     fixedTime.Add(90 * time.Hour),
		Description:          "Updated description",
		MultipleProviders:    false,
		MinAmountPerProvider: 5,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		seed      *domain.SupplyRequest
		req       dto.SupplyGeneralUpdateDTO
		updateErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, wantErr: auth.ErrUnauthenticated},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, req: validReq, wantErr: domain.ErrInvalidInput},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), req: validReq, wantErr: domain.ErrNotFound},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, req: validReq, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, req: validReq, wantErr: domain.ErrInvalidRequestStatus},
		{name: "empty product name", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: dto.SupplyGeneralUpdateDTO{TotalAmount: 80, ActualAmount: 60}, wantErr: domain.ErrInvalidInput},
		{name: "actual amount above total", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: dto.SupplyGeneralUpdateDTO{ProductName: "Maize", TotalAmount: 10, ActualAmount: 20}, wantErr: domain.ErrInvalidInput},
		{
			name: "deadline order invalid",
			ctx:  principalCtx(),
			id:   ownedRequest.ID,
			seed: &ownedRequest,
			req: dto.SupplyGeneralUpdateDTO{
				ProductName:      "Maize",
				TotalAmount:      80,
				ActualAmount:     80,
				RequestDeadline:  fixedTime.Add(90 * time.Hour),
				DeliveryDeadline: fixedTime.Add(30 * time.Hour),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.updateErr = tt.updateErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			err := uc.Update(tt.ctx, tt.id, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(requestRepo.updated) != 1 {
				t.Fatalf("repo updates = %d, want 1", len(requestRepo.updated))
			}
			saved := requestRepo.requests[tt.id]
			if saved.ProductName != "Maize" {
				t.Errorf("product name = %q, want Maize", saved.ProductName)
			}
			if saved.TotalAmount != 80 || saved.ActualAmount != 60 {
				t.Errorf("amounts = %v / %v, want 80 / 60", saved.TotalAmount, saved.ActualAmount)
			}
			if saved.AmountUnit != domain.Lb || saved.UnitOfMeasure != domain.Lb {
				t.Errorf("units = %v / %v, want Lb / Lb", saved.AmountUnit, saved.UnitOfMeasure)
			}
			if saved.MultipleProviders || saved.MinAmountPerProvider != 5 {
				t.Errorf("provider policy = %v / %v, want false / 5", saved.MultipleProviders, saved.MinAmountPerProvider)
			}
			if !saved.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", saved.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestSupplyRequestUseCaseUpdateAmounts(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)
	cancelledRequest := supplyTestRequest(testUserID)
	cancelledRequest.Status = domain.SupplyRequestCancelled

	validReq := dto.SupplyUpdateAmountsDTO{
		TotalAmount:          120,
		ActualAmount:         90,
		AmountUnit:           domain.Tn,
		AmountPerUnit:        15,
		UnitOfMeasure:        domain.Tn,
		MultipleProviders:    true,
		MinAmountPerProvider: 20,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		seed      *domain.SupplyRequest
		req       dto.SupplyUpdateAmountsDTO
		updateErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, wantErr: auth.ErrUnauthenticated},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, req: validReq, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, req: validReq, wantErr: domain.ErrInvalidRequestStatus},
		{name: "actual amount above total", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: dto.SupplyUpdateAmountsDTO{TotalAmount: 10, ActualAmount: 50}, wantErr: domain.ErrInvalidInput},
		{name: "zero total amount", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: dto.SupplyUpdateAmountsDTO{TotalAmount: 0}, wantErr: domain.ErrInvalidInput},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.updateErr = tt.updateErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			err := uc.UpdateAmounts(tt.ctx, tt.id, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := requestRepo.requests[tt.id]
			if saved.TotalAmount != 120 || saved.ActualAmount != 90 {
				t.Errorf("amounts = %v / %v, want 120 / 90", saved.TotalAmount, saved.ActualAmount)
			}
			if saved.AmountUnit != domain.Tn || saved.UnitOfMeasure != domain.Tn {
				t.Errorf("units = %v / %v, want Tn / Tn", saved.AmountUnit, saved.UnitOfMeasure)
			}
			if saved.MinAmountPerProvider != 20 {
				t.Errorf("min amount per provider = %v, want 20", saved.MinAmountPerProvider)
			}
			if !saved.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", saved.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestSupplyRequestUseCaseUpdateRespectsMatchedAmount(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)

	baseReq := func(totalAmount, actualAmount float64) dto.SupplyGeneralUpdateDTO {
		return dto.SupplyGeneralUpdateDTO{
			ProductName:      "Rice",
			TotalAmount:      totalAmount,
			ActualAmount:     actualAmount,
			RequestDeadline:  fixedTime.Add(24 * time.Hour),
			DeliveryDeadline: fixedTime.Add(72 * time.Hour),
		}
	}

	tests := []struct {
		name          string
		req           dto.SupplyGeneralUpdateDTO
		matched       float64
		listActiveErr error
		wantErr       error
		wantTotal     float64
		wantActual    float64
	}{
		{name: "total below matched amount", req: baseReq(30, 0), matched: 40, wantErr: domain.ErrInsufficientAmount},
		{name: "total equal to matched amount", req: baseReq(40, 0), matched: 40, wantTotal: 40, wantActual: 0},
		{name: "actual above total minus matched", req: baseReq(100, 70), matched: 40, wantErr: domain.ErrInsufficientAmount},
		{name: "no active matches keeps behaviour", req: baseReq(100, 60), wantTotal: 100, wantActual: 60},
		{name: "match repo error", req: baseReq(100, 60), wantErr: errFake, listActiveErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			requestRepo.requests[ownedRequest.ID] = ownedRequest
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.activeByRequest = supplyMatchedMatches(ownedRequest.ID, tt.matched)
			matchRepo.listActiveByRequestErr = tt.listActiveErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := uc.Update(principalCtx(), ownedRequest.ID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(requestRepo.updated) != 1 {
				t.Fatalf("repo updates = %d, want 1", len(requestRepo.updated))
			}
			saved := requestRepo.requests[ownedRequest.ID]
			if saved.TotalAmount != tt.wantTotal || saved.ActualAmount != tt.wantActual {
				t.Errorf("amounts = %v / %v, want %v / %v", saved.TotalAmount, saved.ActualAmount, tt.wantTotal, tt.wantActual)
			}
		})
	}
}

func TestSupplyRequestUseCaseUpdateAmountsRespectsMatchedAmount(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)

	baseReq := func(totalAmount, actualAmount float64) dto.SupplyUpdateAmountsDTO {
		return dto.SupplyUpdateAmountsDTO{
			TotalAmount:   totalAmount,
			ActualAmount:  actualAmount,
			AmountUnit:    domain.Kg,
			AmountPerUnit: 10,
			UnitOfMeasure: domain.Kg,
		}
	}

	tests := []struct {
		name          string
		req           dto.SupplyUpdateAmountsDTO
		matched       float64
		listActiveErr error
		wantErr       error
		wantTotal     float64
		wantActual    float64
	}{
		{name: "total below matched amount", req: baseReq(30, 0), matched: 40, wantErr: domain.ErrInsufficientAmount},
		{name: "total equal to matched amount", req: baseReq(40, 0), matched: 40, wantTotal: 40, wantActual: 0},
		{name: "actual above total minus matched", req: baseReq(100, 70), matched: 40, wantErr: domain.ErrInsufficientAmount},
		{name: "no active matches keeps behaviour", req: baseReq(100, 60), wantTotal: 100, wantActual: 60},
		{name: "match repo error", req: baseReq(100, 60), listActiveErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			requestRepo.requests[ownedRequest.ID] = ownedRequest
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.activeByRequest = supplyMatchedMatches(ownedRequest.ID, tt.matched)
			matchRepo.listActiveByRequestErr = tt.listActiveErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := uc.UpdateAmounts(principalCtx(), ownedRequest.ID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(requestRepo.updated) != 1 {
				t.Fatalf("repo updates = %d, want 1", len(requestRepo.updated))
			}
			saved := requestRepo.requests[ownedRequest.ID]
			if saved.TotalAmount != tt.wantTotal || saved.ActualAmount != tt.wantActual {
				t.Errorf("amounts = %v / %v, want %v / %v", saved.TotalAmount, saved.ActualAmount, tt.wantTotal, tt.wantActual)
			}
		})
	}
}

func TestSupplyRequestUseCaseUpdateDeadlines(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)
	cancelledRequest := supplyTestRequest(testUserID)
	cancelledRequest.Status = domain.SupplyRequestCancelled

	validReq := dto.SupplyUpdateTimeDTO{
		RequestDeadline:  fixedTime.Add(30 * time.Hour),
		DeliveryDeadline: fixedTime.Add(90 * time.Hour),
	}

	tests := []struct {
		name      string
		ctx       context.Context
		id        uuid.UUID
		seed      *domain.SupplyRequest
		req       dto.SupplyUpdateTimeDTO
		updateErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, wantErr: auth.ErrUnauthenticated},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, req: validReq, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, req: validReq, wantErr: domain.ErrInvalidRequestStatus},
		{
			name: "deadline order invalid",
			ctx:  principalCtx(),
			id:   ownedRequest.ID,
			seed: &ownedRequest,
			req: dto.SupplyUpdateTimeDTO{
				RequestDeadline:  fixedTime.Add(90 * time.Hour),
				DeliveryDeadline: fixedTime.Add(30 * time.Hour),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, req: validReq, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.updateErr = tt.updateErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), newSupplyFakeMatchRepo(), newFakeTimer())

			err := uc.UpdateDeadlines(tt.ctx, tt.id, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := requestRepo.requests[tt.id]
			if !saved.RequestDeadline.Equal(validReq.RequestDeadline) || !saved.DeliveryDeadline.Equal(validReq.DeliveryDeadline) {
				t.Errorf("deadlines = %v / %v, want %v / %v", saved.RequestDeadline, saved.DeliveryDeadline, validReq.RequestDeadline, validReq.DeliveryDeadline)
			}
			if !saved.UpdatedAt.Equal(fixedTime) {
				t.Errorf("updated at = %v, want %v", saved.UpdatedAt, fixedTime)
			}
		})
	}
}

func TestSupplyRequestUseCaseCancel(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)
	cancelledRequest := supplyTestRequest(testUserID)
	cancelledRequest.Status = domain.SupplyRequestCancelled

	tests := []struct {
		name         string
		ctx          context.Context
		id           uuid.UUID
		seed         *domain.SupplyRequest
		existsActive bool
		existsErr    error
		updateErr    error
		wantErr      error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, wantErr: auth.ErrUnauthenticated},
		{name: "null id", ctx: principalCtx(), id: uuid.Nil, wantErr: domain.ErrInvalidInput},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, wantErr: domain.ErrInvalidRequestStatus},
		{name: "active match blocks cancel", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, existsActive: true, wantErr: primary.ErrActiveMatch},
		{name: "match repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, existsErr: errFake, wantErr: errFake},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.updateErr = tt.updateErr
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.existsActive = tt.existsActive
			matchRepo.existsErr = tt.existsErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := uc.Cancel(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := requestRepo.requests[tt.id]
			if saved.Status != domain.SupplyRequestCancelled {
				t.Errorf("status = %v, want cancelled", saved.Status)
			}
		})
	}
}

func TestSupplyRequestUseCaseExpire(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)
	otherRequest := supplyTestRequest(testOtherID)
	cancelledRequest := supplyTestRequest(testUserID)
	cancelledRequest.Status = domain.SupplyRequestCancelled

	tests := []struct {
		name         string
		ctx          context.Context
		id           uuid.UUID
		seed         *domain.SupplyRequest
		existsActive bool
		existsErr    error
		updateErr    error
		wantErr      error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, wantErr: auth.ErrUnauthenticated},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, wantErr: domain.ErrInvalidRequestStatus},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
		{name: "active match blocks expire", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, existsActive: true, wantErr: primary.ErrActiveMatch},
		{name: "match repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, existsErr: errFake, wantErr: errFake},
		{name: "repo error", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest, updateErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			if tt.seed != nil {
				requestRepo.requests[tt.seed.ID] = *tt.seed
			}
			requestRepo.updateErr = tt.updateErr
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.existsActive = tt.existsActive
			matchRepo.existsErr = tt.existsErr
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := uc.Expire(tt.ctx, tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(requestRepo.updated) != 0 {
					t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			saved := requestRepo.requests[tt.id]
			if saved.Status != domain.SupplyRequestExpired {
				t.Errorf("status = %v, want expired", saved.Status)
			}
		})
	}
}

// Expiring a request that still has an active match would strand the
// reservation taken by that match: the request leaves the open set while the
// matched amount stays deducted from actual_amount forever. Expire must refuse
// exactly like Cancel does, and the request must remain Open.
func TestSupplyRequestUseCaseExpireRefusesWhileActiveMatchStrandsReservation(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)

	requestRepo := newSupplyFakeRequestRepo()
	requestRepo.requests[ownedRequest.ID] = ownedRequest
	matchRepo := newSupplyFakeMatchRepo()
	matchRepo.existsActive = true
	uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

	err := uc.Expire(principalCtx(), ownedRequest.ID)
	if !errors.Is(err, primary.ErrActiveMatch) {
		t.Fatalf("error = %v, want %v", err, primary.ErrActiveMatch)
	}

	saved := requestRepo.requests[ownedRequest.ID]
	if saved.Status != domain.SupplyRequestOpen {
		t.Errorf("status = %v, want open", saved.Status)
	}
	if len(requestRepo.updated) != 0 {
		t.Errorf("repo updates = %d, want 0", len(requestRepo.updated))
	}
}

type supplyEditFunc func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error

// RF-10: "Una solicitud no podrá modificarse después de aceptar una oferta."
//
// A request holding an active match is still IsOpen, so the status guard never
// caught this: the buyer could rewrite the product name, the unit of measure,
// the address and the deadlines behind an offer they had already accepted.
// validateAgainstMatchedAmount only constrains the two amounts, so the rule had
// no enforcement at all. Each of the three edit paths now refuses, and refuses
// before the repository is reached.
func TestSupplyRequestUseCaseEditsRefusedOnceOfferAccepted(t *testing.T) {
	t.Parallel()

	updates := []struct {
		name   string
		invoke supplyEditFunc
	}{
		{
			name: "Update",
			invoke: func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
				return uc.Update(ctx, id, dto.SupplyGeneralUpdateDTO{
					ProductName:      "Maize",
					TotalAmount:      80,
					ActualAmount:     60,
					AmountUnit:       domain.Lb,
					NumberOfUnits:    8,
					AmountPerUnit:    10,
					UnitOfMeasure:    domain.Lb,
					Address:          domain.Address{Department: "León", Municipality: "León", AddressLine: "Central Market"},
					RequestDeadline:  fixedTime.Add(30 * time.Hour),
					DeliveryDeadline: fixedTime.Add(90 * time.Hour),
				})
			},
		},
		{
			name: "UpdateAmounts",
			invoke: func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
				return uc.UpdateAmounts(ctx, id, dto.SupplyUpdateAmountsDTO{
					TotalAmount:          80,
					ActualAmount:         60,
					AmountUnit:           domain.Lb,
					AmountPerUnit:        10,
					UnitOfMeasure:        domain.Lb,
					MultipleProviders:    true,
					MinAmountPerProvider: 5,
				})
			},
		},
		{
			name: "UpdateDeadlines",
			invoke: func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
				return uc.UpdateDeadlines(ctx, id, dto.SupplyUpdateTimeDTO{
					RequestDeadline:  fixedTime.Add(30 * time.Hour),
					DeliveryDeadline: fixedTime.Add(90 * time.Hour),
				})
			},
		},
	}

	tests := []struct {
		name        string
		activeMatch bool
		wantErr     error
	}{
		{name: "offer accepted", activeMatch: true, wantErr: primary.ErrActiveMatch},
		{name: "no offer accepted"},
	}

	for _, update := range updates {
		for _, tt := range tests {
			t.Run(update.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				ownedRequest := supplyTestRequest(testUserID)

				requestRepo := newSupplyFakeRequestRepo()
				requestRepo.requests[ownedRequest.ID] = ownedRequest
				matchRepo := newSupplyFakeMatchRepo()
				matchRepo.existsActive = tt.activeMatch
				uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

				err := update.invoke(uc, principalCtx(), ownedRequest.ID)

				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("error = %v, want %v", err, tt.wantErr)
					}
					// Refused before persistence: the row is untouched.
					if len(requestRepo.updated) != 0 {
						t.Fatalf("repo updates = %d, want 0", len(requestRepo.updated))
					}
					saved := requestRepo.requests[ownedRequest.ID]
					if saved.ProductName != "Rice" || saved.TotalAmount != 100 {
						t.Errorf("stored request = %q / %v, want the untouched Rice / 100", saved.ProductName, saved.TotalAmount)
					}
					if !saved.RequestDeadline.Equal(ownedRequest.RequestDeadline) || !saved.DeliveryDeadline.Equal(ownedRequest.DeliveryDeadline) {
						t.Errorf("deadlines = %v / %v, want the untouched %v / %v",
							saved.RequestDeadline, saved.DeliveryDeadline, ownedRequest.RequestDeadline, ownedRequest.DeliveryDeadline)
					}
					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(requestRepo.updated) != 1 {
					t.Fatalf("repo updates = %d, want 1", len(requestRepo.updated))
				}
			})
		}
	}
}

// A match repository failure is not an answer to "is there an active offer?".
// The edits propagate it rather than assuming there is none and writing.
func TestSupplyRequestUseCaseEditsPropagateMatchLookupFailure(t *testing.T) {
	t.Parallel()

	ownedRequest := supplyTestRequest(testUserID)

	invocations := map[string]supplyEditFunc{
		"Update": func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
			return uc.Update(ctx, id, dto.SupplyGeneralUpdateDTO{
				ProductName:     "Maize",
				TotalAmount:     80,
				ActualAmount:    60,
				RequestDeadline: fixedTime.Add(30 * time.Hour),
			})
		},
		"UpdateAmounts": func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
			return uc.UpdateAmounts(ctx, id, dto.SupplyUpdateAmountsDTO{TotalAmount: 80, ActualAmount: 60})
		},
		"UpdateDeadlines": func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
			return uc.UpdateDeadlines(ctx, id, dto.SupplyUpdateTimeDTO{
				RequestDeadline:  fixedTime.Add(30 * time.Hour),
				DeliveryDeadline: fixedTime.Add(90 * time.Hour),
			})
		},
	}

	for name, invoke := range invocations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requestRepo := newSupplyFakeRequestRepo()
			requestRepo.requests[ownedRequest.ID] = ownedRequest
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.existsErr = errFake
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := invoke(uc, principalCtx(), ownedRequest.ID)

			if !errors.Is(err, errFake) {
				t.Fatalf("error = %v, want %v", err, errFake)
			}
			if len(requestRepo.updated) != 0 {
				t.Errorf("repo updates = %d, want 0", len(requestRepo.updated))
			}
		})
	}
}

// Cancel and Expire are guarded by the same predicate and that is deliberate:
// RF-10 lets the buyer cancel "siempre que aún no haya aceptado una oferta", so
// refusing to cancel while a match exists is the rule, not an oversight. The
// exit path is cancelling the individual transaction, not the request. This
// test exists so a future reader who spots the duplication does not "fix" it by
// relaxing Cancel and Expire to match the edits.
func TestSupplyRequestUseCaseCancelAndExpireStayRefusedOnceOfferAccepted(t *testing.T) {
	t.Parallel()

	invocations := map[string]supplyEditFunc{
		"Cancel": func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
			return uc.Cancel(ctx, id)
		},
		"Expire": func(uc *usecases.SupplyRequestUseCaseImpl, ctx context.Context, id uuid.UUID) error {
			return uc.Expire(ctx, id)
		},
	}

	for name, invoke := range invocations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ownedRequest := supplyTestRequest(testUserID)

			requestRepo := newSupplyFakeRequestRepo()
			requestRepo.requests[ownedRequest.ID] = ownedRequest
			matchRepo := newSupplyFakeMatchRepo()
			matchRepo.existsActive = true
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeOfferRepo(), matchRepo, newFakeTimer())

			err := invoke(uc, principalCtx(), ownedRequest.ID)

			if !errors.Is(err, primary.ErrActiveMatch) {
				t.Fatalf("error = %v, want %v", err, primary.ErrActiveMatch)
			}
			if len(requestRepo.updated) != 0 {
				t.Errorf("repo updates = %d, want 0", len(requestRepo.updated))
			}
			saved := requestRepo.requests[ownedRequest.ID]
			if saved.Status != domain.SupplyRequestOpen {
				t.Errorf("status = %v, want open", saved.Status)
			}
		})
	}
}
