package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/internal/auth"
)

var testE2EBuyerID uuid.UUID = uuid.MustParse("e2e00000-0000-4000-8000-000000000001")
var testE2ESupplierID uuid.UUID = uuid.MustParse("e2e00000-0000-4000-8000-000000000002")

// e2eOfferPrice is the unit price the end-to-end supplier quotes. RF-11 makes
// the price a required field, so the fixture carries one; nothing reads it.
var e2eOfferPrice = 25.5

func TestE2ESupplyRequestToCompletedTransaction(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	clock := timepkg.NewClock()

	userRepo := repository.NewUserRepository(TestPool)
	requestRepo := repository.NewSupplyRequestRepository(TestPool)
	offerRepo := repository.NewSupplyOfferRepository(TestPool)
	matchRepo := repository.NewMatchRepository(TestPool)
	transactionRepo := repository.NewTransactionRepository(TestPool)
	inventoryRepo := repository.NewSupplierInventoryRepository(TestPool)
	unitOfWork := repository.NewUnitOfWork(TestPool)

	buyer := &domain.User{
		ID:           testE2EBuyerID,
		FirstName:    "E2E",
		LastName:     "Buyer",
		Role:         domain.RoleCompradorMayoristaDetallista,
		Email:        "e2e-buyer@example.com",
		PhoneNumber:  "4600-0001",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}
	supplier := &domain.User{
		ID:           testE2ESupplierID,
		FirstName:    "E2E",
		LastName:     "Supplier",
		Role:         domain.RoleAgricultor,
		Email:        "e2e-supplier@example.com",
		PhoneNumber:  "4600-0002",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}
	for _, u := range []*domain.User{buyer, supplier} {
		if _, err := userRepo.Save(ctx, u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}

	inventory := &domain.SupplierInventory{
		ID:          uuid.New(),
		SupplierID:  testE2ESupplierID,
		ProductName: "Maize",
		Quantity:    500,
		AmountUnit:  domain.Kg,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("insert fixture supplier inventory: %v", err)
	}

	buyerCtx := auth.WithPrincipal(ctx, auth.Principal{UserID: testE2EBuyerID, Role: domain.RoleCompradorMayoristaDetallista})
	supplierCtx := auth.WithPrincipal(ctx, auth.Principal{UserID: testE2ESupplierID, Role: domain.RoleAgricultor})

	supplyRequestUC := usecases.NewSupplyRequestUseCase(requestRepo, offerRepo, matchRepo, clock)
	supplyOfferUC := usecases.NewSupplyOfferUseCase(offerRepo, requestRepo, matchRepo, clock)
	recommendationUC := usecases.NewRecommendationUseCase(offerRepo, requestRepo, userRepo, inventoryRepo, matchRepo, nil)
	matchUC := usecases.NewMatchUseCase(requestRepo, offerRepo, matchRepo, transactionRepo, recommendationUC, unitOfWork)
	transactionUC := usecases.NewTransactionUseCase(transactionRepo, matchRepo, requestRepo, offerRepo, clock, unitOfWork)

	totalAmount := float64(100)

	createdRequest, err := supplyRequestUC.Create(buyerCtx, dto.SupplyRequestDTO{
		ProductName:       "Maize",
		TotalAmount:       totalAmount,
		AmountUnit:        domain.Kg,
		NumberOfUnits:     10,
		AmountPerUnit:     10,
		UnitOfMeasure:     domain.Kg,
		Address:           domain.Address{Department: "Masaya", Municipality: "Masaya", AddressLine: "Km 5 Carretera Sur"},
		RequestDeadline:   time.Now().Add(24 * time.Hour),
		DeliveryDeadline:  time.Now().Add(48 * time.Hour),
		Description:       "e2e supply flow",
		MultipleProviders: false,
	})
	if err != nil {
		t.Fatalf("SupplyRequestUseCase.Create() error: %v", err)
	}
	if createdRequest.ID == nil || *createdRequest.ID == uuid.Nil {
		t.Fatal("SupplyRequestUseCase.Create() returned nil id")
	}
	if createdRequest.Status != domain.SupplyRequestOpen {
		t.Errorf("request status = %v, want %v", createdRequest.Status, domain.SupplyRequestOpen)
	}
	if createdRequest.ActualAmount != createdRequest.TotalAmount {
		t.Errorf("request ActualAmount = %v, want TotalAmount %v", createdRequest.ActualAmount, createdRequest.TotalAmount)
	}
	requestID := *createdRequest.ID

	createdOffer, err := supplyOfferUC.Create(supplierCtx, dto.SupplyOfferDTO{
		SupplyRequest:       &requestID,
		TotalAmount:         totalAmount,
		AmountUnit:          domain.Kg,
		PricePerUnit:        &e2eOfferPrice,
		ProposedDeliveryDay: time.Now().Add(72 * time.Hour),
		DeliveryAvailable:   true,
	})
	if err != nil {
		t.Fatalf("SupplyOfferUseCase.Create() error: %v", err)
	}
	if createdOffer.ID == nil || *createdOffer.ID == uuid.Nil {
		t.Fatal("SupplyOfferUseCase.Create() returned nil id")
	}
	offerID := *createdOffer.ID

	offerAfterCreate, err := offerRepo.GetByID(ctx, offerID)
	if err != nil {
		t.Fatalf("offer GetByID() after Create error: %v", err)
	}
	if offerAfterCreate.Status != domain.OfferActive {
		t.Errorf("offer status after create = %v, want %v", offerAfterCreate.Status, domain.OfferActive)
	}

	matchDTO, transactionDTO, err := matchUC.Like(buyerCtx, offerID)
	if err != nil {
		t.Fatalf("MatchUseCase.Like() error: %v", err)
	}

	matchAfterLike, err := matchRepo.GetByID(ctx, matchDTO.ID)
	if err != nil {
		t.Fatalf("match GetByID() after Like error: %v", err)
	}
	if matchAfterLike.Status != domain.MatchActive {
		t.Errorf("match status = %v, want %v", matchAfterLike.Status, domain.MatchActive)
	}
	if matchAfterLike.MatchedAmount != totalAmount {
		t.Errorf("match MatchedAmount = %v, want %v", matchAfterLike.MatchedAmount, totalAmount)
	}
	if matchAfterLike.SupplyOffer != offerID || matchAfterLike.SupplyRequest != requestID {
		t.Errorf("match links = offer %v request %v, want offer %v request %v",
			matchAfterLike.SupplyOffer, matchAfterLike.SupplyRequest, offerID, requestID)
	}

	offerAfterLike, err := offerRepo.GetByID(ctx, offerID)
	if err != nil {
		t.Fatalf("offer GetByID() after Like error: %v", err)
	}
	if offerAfterLike.Status != domain.OfferMatched {
		t.Errorf("offer status after like = %v, want %v", offerAfterLike.Status, domain.OfferMatched)
	}

	requestAfterLike, err := requestRepo.GetByID(ctx, requestID)
	if err != nil {
		t.Fatalf("request GetByID() after Like error: %v", err)
	}
	if want := totalAmount - matchAfterLike.MatchedAmount; requestAfterLike.ActualAmount != want {
		t.Errorf("request ActualAmount after like = %v, want %v", requestAfterLike.ActualAmount, want)
	}
	if requestAfterLike.Status != domain.SupplyRequestOpen {
		t.Errorf("request status after like = %v, want %v", requestAfterLike.Status, domain.SupplyRequestOpen)
	}

	transactionAfterLike, err := transactionRepo.GetByMatch(ctx, matchDTO.ID)
	if err != nil {
		t.Fatalf("transaction GetByMatch() after Like error: %v", err)
	}
	if transactionAfterLike.Status != domain.TransactionMatched {
		t.Errorf("transaction status after like = %v, want %v", transactionAfterLike.Status, domain.TransactionMatched)
	}
	transactionID := transactionAfterLike.ID
	if transactionDTO.ID == nil || *transactionDTO.ID != transactionID {
		t.Errorf("Like() transaction id = %v, want %v", transactionDTO.ID, transactionID)
	}

	if err := transactionUC.ConfirmStart(buyerCtx, transactionID); err != nil {
		t.Fatalf("ConfirmStart(buyer) error: %v", err)
	}
	afterFirstStart, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("transaction GetByID() after first ConfirmStart error: %v", err)
	}
	if afterFirstStart.Status != domain.TransactionMatched {
		t.Errorf("transaction status after first ConfirmStart = %v, want %v", afterFirstStart.Status, domain.TransactionMatched)
	}
	if afterFirstStart.BuyerStartConfirmedAt == nil {
		t.Error("transaction BuyerStartConfirmedAt is nil after buyer ConfirmStart")
	}
	if afterFirstStart.SupplierStartConfirmedAt != nil {
		t.Error("transaction SupplierStartConfirmedAt is set before supplier ConfirmStart")
	}

	if err := transactionUC.ConfirmStart(supplierCtx, transactionID); err != nil {
		t.Fatalf("ConfirmStart(supplier) error: %v", err)
	}
	afterBothStarts, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("transaction GetByID() after both ConfirmStart error: %v", err)
	}
	if afterBothStarts.Status != domain.TransactionInProgress {
		t.Errorf("transaction status after both ConfirmStart = %v, want %v", afterBothStarts.Status, domain.TransactionInProgress)
	}
	if afterBothStarts.BuyerStartConfirmedAt == nil || afterBothStarts.SupplierStartConfirmedAt == nil {
		t.Error("transaction start confirmations missing after both ConfirmStart calls")
	}

	if err := transactionUC.ConfirmDelivery(buyerCtx, transactionID); err != nil {
		t.Fatalf("ConfirmDelivery(buyer) error: %v", err)
	}
	afterFirstDelivery, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("transaction GetByID() after first ConfirmDelivery error: %v", err)
	}
	if afterFirstDelivery.Status != domain.TransactionInProgress {
		t.Errorf("transaction status after first ConfirmDelivery = %v, want %v", afterFirstDelivery.Status, domain.TransactionInProgress)
	}
	if afterFirstDelivery.BuyerDeliveryConfirmedAt == nil {
		t.Error("transaction BuyerDeliveryConfirmedAt is nil after buyer ConfirmDelivery")
	}

	if err := transactionUC.ConfirmDelivery(supplierCtx, transactionID); err != nil {
		t.Fatalf("ConfirmDelivery(supplier) error: %v", err)
	}
	afterBothDeliveries, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("transaction GetByID() after both ConfirmDelivery error: %v", err)
	}
	if afterBothDeliveries.Status != domain.TransactionCompleted {
		t.Errorf("transaction status after both ConfirmDelivery = %v, want %v", afterBothDeliveries.Status, domain.TransactionCompleted)
	}

	requestAfterComplete, err := requestRepo.GetByID(ctx, requestID)
	if err != nil {
		t.Fatalf("request GetByID() after ConfirmDelivery error: %v", err)
	}
	if requestAfterComplete.Status != domain.SupplyRequestCompleted {
		t.Errorf("request status after completed transaction = %v, want %v", requestAfterComplete.Status, domain.SupplyRequestCompleted)
	}

	finalTransaction, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("final transaction GetByID() error: %v", err)
	}
	if finalTransaction.Status != domain.TransactionCompleted {
		t.Errorf("final transaction status = %v, want %v", finalTransaction.Status, domain.TransactionCompleted)
	}
	if finalTransaction.MatchID != matchDTO.ID {
		t.Errorf("final transaction MatchID = %v, want %v", finalTransaction.MatchID, matchDTO.ID)
	}
	if finalTransaction.BuyerStartConfirmedAt == nil || finalTransaction.SupplierStartConfirmedAt == nil ||
		finalTransaction.BuyerDeliveryConfirmedAt == nil || finalTransaction.SupplierDeliveryConfirmedAt == nil {
		t.Error("final transaction is missing one or more confirmation timestamps")
	}

	finalMatch, err := matchRepo.GetByID(ctx, matchDTO.ID)
	if err != nil {
		t.Fatalf("final match GetByID() error: %v", err)
	}
	if finalMatch.Status != domain.MatchActive {
		t.Errorf("final match status = %v, want %v", finalMatch.Status, domain.MatchActive)
	}
	if finalMatch.MatchedAmount != totalAmount {
		t.Errorf("final match MatchedAmount = %v, want %v", finalMatch.MatchedAmount, totalAmount)
	}

	finalOffer, err := offerRepo.GetByID(ctx, offerID)
	if err != nil {
		t.Fatalf("final offer GetByID() error: %v", err)
	}
	if finalOffer.Status != domain.OfferMatched {
		t.Errorf("final offer status = %v, want %v", finalOffer.Status, domain.OfferMatched)
	}

	finalRequest, err := requestRepo.GetByID(ctx, requestID)
	if err != nil {
		t.Fatalf("final request GetByID() error: %v", err)
	}
	if finalRequest.Status != domain.SupplyRequestCompleted {
		t.Errorf("final request status = %v, want %v", finalRequest.Status, domain.SupplyRequestCompleted)
	}
	if finalRequest.ActualAmount != 0 {
		t.Errorf("final request ActualAmount = %v, want 0", finalRequest.ActualAmount)
	}
}
