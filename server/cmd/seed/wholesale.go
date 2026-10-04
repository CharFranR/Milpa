package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	repo "milpa/infrastructure/adapters/secondary/repository"
	internalauth "milpa/internal/auth"
)

const (
	demoBuyerEmail          = "mayorista.elroble@milpa.com"
	demoLiquidationProduct  = "Lote de café pergamino"
	demoLiquidationQuantity = 120
)

type wholesaleSeedDeps struct {
	userRepo      *repo.UserRepositoryImpl
	userUC        *usecases.UserUseCaseImpl
	requestRepo   port.SupplyRequestRepository
	requestUC     primary.SupplyRequestUseCase
	offerUC       primary.SupplyOfferUseCase
	matchUC       primary.MatchUseCase
	transactionUC primary.TransactionUseCase
	liquidationUC primary.LiquidationUseCase
	farmerID      uuid.UUID
}

type wholesaleSeedSummary struct {
	BuyerID             uuid.UUID
	BuyerState          string
	RequestsCreated     int
	RequestsSkipped     int
	OffersCreated       int
	DealsCompleted      int
	LiquidationsCreated int
}

type demoWholesaleRequest struct {
	productName          string
	description          string
	totalAmount          float64
	numberOfUnits        float64
	amountPerUnit        float64
	multipleProviders    bool
	minAmountPerProvider float64
	offerAmount          float64
	offerPricePerUnit    float64
	offerComments        string
	completeDeal         bool
	cancelRequest        bool
}

func demoWholesaleRequests() []demoWholesaleRequest {
	return []demoWholesaleRequest{
		{
			productName:          "Café pergamino",
			description:          "Compra de temporada para tueste propio.",
			totalAmount:          800,
			numberOfUnits:        100,
			amountPerUnit:        8,
			multipleProviders:    true,
			minAmountPerProvider: 100,
			offerAmount:          300,
			offerPricePerUnit:    8.5,
			offerComments:        "Cosecha de marzo, secado al sol.",
		},
		{
			productName:          "Maíz blanco",
			description:          "Lote cerrado para distribución en bodega.",
			totalAmount:          400,
			numberOfUnits:        80,
			amountPerUnit:        5,
			multipleProviders:    false,
			minAmountPerProvider: 100,
			offerAmount:          400,
			offerPricePerUnit:    5,
			offerComments:        "Entrega única con transporte incluido.",
			completeDeal:         true,
		},
		{
			productName:          "Frijol rojo",
			description:          "Solicitud cancelada de ejemplo.",
			totalAmount:          250,
			numberOfUnits:        50,
			amountPerUnit:        5,
			multipleProviders:    true,
			minAmountPerProvider: 75,
			offerAmount:          150,
			offerPricePerUnit:    5.2,
			offerComments:        "Disponible a partir de la próxima semana.",
			cancelRequest:        true,
		},
	}
}

func seedWholesale(ctx context.Context, deps wholesaleSeedDeps) (wholesaleSeedSummary, error) {
	summary := wholesaleSeedSummary{BuyerState: "reused"}

	buyerID, created, err := ensureBuyer(ctx, deps.userRepo, deps.userUC)
	if err != nil {
		return summary, fmt.Errorf("ensure demo buyer: %w", err)
	}
	summary.BuyerID = buyerID
	if created {
		summary.BuyerState = "created"
	}

	existing, err := deps.requestRepo.List(ctx, buyerID)
	if err != nil {
		return summary, fmt.Errorf("list demo buyer requests: %w", err)
	}
	seen := make(map[string]struct{}, len(existing))
	for i := range existing {
		seen[existing[i].ProductName] = struct{}{}
	}

	buyerCtx := internalauth.WithPrincipal(ctx, internalauth.Principal{
		UserID: buyerID,
		Role:   domain.RoleCompradorMayoristaDetallista,
	})
	farmerCtx := internalauth.WithPrincipal(ctx, internalauth.Principal{
		UserID: deps.farmerID,
		Role:   domain.RoleAgricultor,
	})

	now := time.Now().UTC()
	for _, item := range demoWholesaleRequests() {
		if _, ok := seen[item.productName]; ok {
			summary.RequestsSkipped++
			continue
		}

		offers, deals, err := seedWholesaleRequest(buyerCtx, farmerCtx, deps, item, now)
		if err != nil {
			return summary, fmt.Errorf("seed supply request %q: %w", item.productName, err)
		}
		summary.RequestsCreated++
		summary.OffersCreated += offers
		summary.DealsCompleted += deals
	}

	liquidations, err := seedWholesaleLiquidation(ctx, farmerCtx, deps, now)
	if err != nil {
		return summary, fmt.Errorf("seed demo liquidation: %w", err)
	}
	summary.LiquidationsCreated = liquidations

	return summary, nil
}

func seedWholesaleRequest(
	buyerCtx context.Context,
	farmerCtx context.Context,
	deps wholesaleSeedDeps,
	item demoWholesaleRequest,
	now time.Time,
) (int, int, error) {
	request, err := deps.requestUC.Create(buyerCtx, dto.SupplyRequestDTO{
		ProductName:          item.productName,
		TotalAmount:          item.totalAmount,
		AmountUnit:           domain.Kg,
		NumberOfUnits:        item.numberOfUnits,
		AmountPerUnit:        item.amountPerUnit,
		UnitOfMeasure:        domain.Kg,
		Address:              demoBuyerAddress(),
		RequestDeadline:      now.AddDate(0, 0, 30),
		DeliveryDeadline:     now.AddDate(0, 0, 60),
		Description:          item.description,
		MultipleProviders:    item.multipleProviders,
		MinAmountPerProvider: item.minAmountPerProvider,
	})
	if err != nil {
		return 0, 0, err
	}

	pricePerUnit := item.offerPricePerUnit
	offer, err := deps.offerUC.Create(farmerCtx, dto.SupplyOfferDTO{
		SupplyRequest:       request.ID,
		TotalAmount:         item.offerAmount,
		AmountUnit:          domain.Kg,
		PricePerUnit:        &pricePerUnit,
		Comments:            item.offerComments,
		ProposedDeliveryDay: now.AddDate(0, 0, 45),
		DeliveryAvailable:   true,
	})
	if err != nil {
		return 0, 0, err
	}

	if item.cancelRequest {
		if err := deps.requestUC.Cancel(buyerCtx, *request.ID); err != nil {
			return 1, 0, err
		}
		return 1, 0, nil
	}

	if !item.completeDeal {
		return 1, 0, nil
	}

	_, transaction, err := deps.matchUC.Like(buyerCtx, *offer.ID)
	if err != nil {
		return 1, 0, err
	}
	transactionID := *transaction.ID

	for _, participant := range []context.Context{buyerCtx, farmerCtx} {
		if err := deps.transactionUC.ConfirmStart(participant, transactionID); err != nil {
			return 1, 0, err
		}
	}
	for _, participant := range []context.Context{buyerCtx, farmerCtx} {
		if err := deps.transactionUC.ConfirmDelivery(participant, transactionID); err != nil {
			return 1, 0, err
		}
	}

	return 1, 1, nil
}

func seedWholesaleLiquidation(ctx context.Context, farmerCtx context.Context, deps wholesaleSeedDeps, now time.Time) (int, error) {
	existing, err := deps.liquidationUC.GetBySupplier(farmerCtx, deps.farmerID)
	if err != nil {
		return 0, err
	}
	for _, liquidation := range existing {
		if liquidation.ProductName == demoLiquidationProduct {
			return 0, nil
		}
	}

	locationID, err := demoFarmerLocation(ctx, deps)
	if err != nil {
		return 0, err
	}

	expiresAt := now.AddDate(0, 0, 20)
	if _, err := deps.liquidationUC.CreateLiquidation(farmerCtx, dto.CreateLiquidationRequest{
		ProductName:   demoLiquidationProduct,
		Quantity:      demoLiquidationQuantity,
		UnitOfMeasure: "quintal",
		TotalPrice:    2160,
		UnitPrice:     18,
		DeliveryTime:  "Dos semanas",
		LocationID:    locationID,
		Visibility:    "public",
		ExpiresAt:     &expiresAt,
	}); err != nil {
		return 0, err
	}

	return 1, nil
}

func demoFarmerLocation(ctx context.Context, deps wholesaleSeedDeps) (uuid.UUID, error) {
	farmer, err := deps.userRepo.FindByID(ctx, deps.farmerID)
	if err != nil {
		return uuid.Nil, err
	}
	if farmer.Address.ID != uuid.Nil {
		return farmer.Address.ID, nil
	}
	return deps.farmerID, nil
}

func ensureBuyer(ctx context.Context, userRepo *repo.UserRepositoryImpl, userUC *usecases.UserUseCaseImpl) (uuid.UUID, bool, error) {
	user, err := userRepo.FindByEmail(ctx, demoBuyerEmail)
	if err == nil {
		return user.ID, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return uuid.Nil, false, err
	}

	registered, err := userUC.Register(ctx, dto.RegisterUserRequest{
		Email:           demoBuyerEmail,
		FirstName:       "Óscar",
		LastName:        "Ramírez",
		Role:            domain.RoleCompradorMayoristaDetallista,
		Address:         "Bodega 12, mercado Mayoreo",
		Department:      "Managua",
		Municipality:    "Managua",
		PhoneNumber:     "+505 8888 0002",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	if err != nil {
		return uuid.Nil, false, err
	}

	return registered.ID, true, nil
}

func demoBuyerAddress() domain.Address {
	return domain.Address{
		Department:   "Managua",
		Municipality: "Managua",
		AddressLine:  "Bodega 12, mercado Mayoreo",
		Latitude:     12.136,
		Longitude:    -86.251,
	}
}
