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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, matchRepo, newFakeTimer())

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
		name      string
		ctx       context.Context
		id        uuid.UUID
		seed      *domain.SupplyRequest
		updateErr error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), id: ownedRequest.ID, seed: &ownedRequest},
		{name: "unauthenticated", ctx: context.Background(), id: ownedRequest.ID, seed: &ownedRequest, wantErr: auth.ErrUnauthenticated},
		{name: "non-owner", ctx: principalCtx(), id: otherRequest.ID, seed: &otherRequest, wantErr: domain.ErrForbidden},
		{name: "not open", ctx: principalCtx(), id: cancelledRequest.ID, seed: &cancelledRequest, wantErr: domain.ErrInvalidRequestStatus},
		{name: "not found", ctx: principalCtx(), id: uuid.New(), wantErr: domain.ErrNotFound},
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
			uc := usecases.NewSupplyRequestUseCase(requestRepo, newSupplyFakeMatchRepo(), newFakeTimer())

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
