package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

var testTransactionSupplierID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d401")
var testTransactionOtherSupplierID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d402")
var testTransactionBuyerID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d403")

var testTransactionRequestID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d411")
var testTransactionRequestID2 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d412")

var testTransactionOfferID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d421")
var testTransactionOfferID2 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d422")
var testTransactionOfferID3 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d423")
var testTransactionOfferID4 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d424")

var testTransactionMatchID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d431")
var testTransactionMatchID2 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d432")
var testTransactionMatchID3 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d433")
var testTransactionMatchID4 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d434")

var testTransactionID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d441")
var testTransactionID2 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d442")
var testTransactionID3 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d443")
var testTransactionID4 uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d444")
var testTransactionNotFoundID uuid.UUID = uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d449")

func newTransactionFixture(id, matchID uuid.UUID, createdAt time.Time) *domain.Transaction {
	return &domain.Transaction{
		ID:           id,
		MatchID:      matchID,
		Status:       domain.TransactionMatched,
		CancelReason: "",
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}
}

func setupTransactionTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)
	requestRepo := repository.NewSupplyRequestRepository(TestPool)
	offerRepo := repository.NewSupplyOfferRepository(TestPool)
	matchRepo := repository.NewMatchRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testTransactionSupplierID,
			FirstName:    "Transaction",
			LastName:     "Supplier",
			Role:         domain.RoleMIPYME,
			Email:        "transaction-supplier@example.com",
			PhoneNumber:  "4400-0001",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testTransactionOtherSupplierID,
			FirstName:    "Transaction",
			LastName:     "OtherSupplier",
			Role:         domain.RoleMIPYME,
			Email:        "transaction-other-supplier@example.com",
			PhoneNumber:  "4400-0002",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testTransactionBuyerID,
			FirstName:    "Transaction",
			LastName:     "Buyer",
			Role:         domain.RoleMIPYME,
			Email:        "transaction-buyer@example.com",
			PhoneNumber:  "4400-0003",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}

	for _, u := range users {
		if _, err := userRepo.Save(context.Background(), u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}

	requests := []*domain.SupplyRequest{
		newSupplyOfferRequestFixture(testTransactionRequestID, testTransactionBuyerID, "Maize"),
		newSupplyOfferRequestFixture(testTransactionRequestID2, testTransactionBuyerID, "Beans"),
	}

	for _, supplyRequest := range requests {
		if err := requestRepo.Create(context.Background(), supplyRequest); err != nil {
			t.Fatalf("insert fixture supply request %s: %v", supplyRequest.ID, err)
		}
	}

	offers := []*domain.SupplyOffer{
		newSupplyOfferFixture(testTransactionOfferID, testTransactionSupplierID, testTransactionRequestID, fixedTime),
		newSupplyOfferFixture(testTransactionOfferID2, testTransactionOtherSupplierID, testTransactionRequestID, fixedTime.Add(time.Minute)),
		newSupplyOfferFixture(testTransactionOfferID3, testTransactionSupplierID, testTransactionRequestID2, fixedTime.Add(2*time.Minute)),
		newSupplyOfferFixture(testTransactionOfferID4, testTransactionOtherSupplierID, testTransactionRequestID2, fixedTime.Add(3*time.Minute)),
	}

	for _, offer := range offers {
		if err := offerRepo.Create(context.Background(), offer); err != nil {
			t.Fatalf("insert fixture supply offer %s: %v", offer.ID, err)
		}
	}

	matches := []*domain.Match{
		newMatchFixture(testTransactionMatchID, testTransactionOfferID, testTransactionRequestID, fixedTime),
		newMatchFixture(testTransactionMatchID2, testTransactionOfferID2, testTransactionRequestID, fixedTime.Add(time.Minute)),
		newMatchFixture(testTransactionMatchID3, testTransactionOfferID3, testTransactionRequestID2, fixedTime.Add(2*time.Minute)),
		newMatchFixture(testTransactionMatchID4, testTransactionOfferID4, testTransactionRequestID2, fixedTime.Add(3*time.Minute)),
	}

	for _, match := range matches {
		if err := matchRepo.Create(context.Background(), match); err != nil {
			t.Fatalf("insert fixture match %s: %v", match.ID, err)
		}
	}
}

func TestTransactionCreateAndGetByID(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	matched := newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime)
	if err := db.Create(context.Background(), matched); err != nil {
		t.Fatalf("Create() matched error: %v", err)
	}
	if matched.ID != testTransactionID {
		t.Errorf("Create() mutated ID = %v, want %v", matched.ID, testTransactionID)
	}

	cancelled := newTransactionFixture(testTransactionID2, testTransactionMatchID2, fixedTime.Add(time.Minute))
	cancelled.Status = domain.TransactionCancelled
	cancelledBy := testTransactionSupplierID
	cancelled.CancelledBy = &cancelledBy
	cancelled.CancelReason = "supplier withdrew"
	if err := db.Create(context.Background(), cancelled); err != nil {
		t.Fatalf("Create() cancelled error: %v", err)
	}

	inProgress := newTransactionFixture(testTransactionID3, testTransactionMatchID3, fixedTime.Add(2*time.Minute))
	inProgress.Status = domain.TransactionInProgress
	buyerStart := fixedTime.Add(30 * time.Minute)
	supplierStart := fixedTime.Add(31 * time.Minute)
	buyerDelivery := fixedTime.Add(40 * time.Minute)
	supplierDelivery := fixedTime.Add(41 * time.Minute)
	inProgress.BuyerStartConfirmedAt = &buyerStart
	inProgress.SupplierStartConfirmedAt = &supplierStart
	inProgress.BuyerDeliveryConfirmedAt = &buyerDelivery
	inProgress.SupplierDeliveryConfirmedAt = &supplierDelivery
	if err := db.Create(context.Background(), inProgress); err != nil {
		t.Fatalf("Create() in progress error: %v", err)
	}

	t.Run("Matched transaction round trip", func(t *testing.T) {
		got, err := db.GetByID(context.Background(), testTransactionID)
		if err != nil {
			t.Fatalf("GetByID() unexpected error: %v", err)
		}

		if got.ID != matched.ID {
			t.Errorf("GetByID() ID = %v, want %v", got.ID, matched.ID)
		}
		if got.MatchID != matched.MatchID {
			t.Errorf("GetByID() MatchID = %v, want %v", got.MatchID, matched.MatchID)
		}
		if got.Status != domain.TransactionMatched {
			t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.TransactionMatched)
		}
		if got.BuyerStartConfirmedAt != nil {
			t.Errorf("GetByID() BuyerStartConfirmedAt = %v, want nil", got.BuyerStartConfirmedAt)
		}
		if got.SupplierStartConfirmedAt != nil {
			t.Errorf("GetByID() SupplierStartConfirmedAt = %v, want nil", got.SupplierStartConfirmedAt)
		}
		if got.BuyerDeliveryConfirmedAt != nil {
			t.Errorf("GetByID() BuyerDeliveryConfirmedAt = %v, want nil", got.BuyerDeliveryConfirmedAt)
		}
		if got.SupplierDeliveryConfirmedAt != nil {
			t.Errorf("GetByID() SupplierDeliveryConfirmedAt = %v, want nil", got.SupplierDeliveryConfirmedAt)
		}
		if got.CancelledBy != nil {
			t.Errorf("GetByID() CancelledBy = %v, want nil", got.CancelledBy)
		}
		if got.CancelReason != "" {
			t.Errorf("GetByID() CancelReason = %q, want empty", got.CancelReason)
		}
		if !got.CreatedAt.Equal(matched.CreatedAt) {
			t.Errorf("GetByID() CreatedAt = %v, want %v", got.CreatedAt, matched.CreatedAt)
		}
		if !got.UpdatedAt.Equal(matched.UpdatedAt) {
			t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, matched.UpdatedAt)
		}
	})

	t.Run("Cancelled transaction round trip", func(t *testing.T) {
		got, err := db.GetByID(context.Background(), testTransactionID2)
		if err != nil {
			t.Fatalf("GetByID() unexpected error: %v", err)
		}

		if got.Status != domain.TransactionCancelled {
			t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.TransactionCancelled)
		}
		if got.CancelledBy == nil {
			t.Fatal("GetByID() CancelledBy = nil, want supplier id")
		}
		if *got.CancelledBy != cancelledBy {
			t.Errorf("GetByID() CancelledBy = %v, want %v", *got.CancelledBy, cancelledBy)
		}
		if got.CancelReason != "supplier withdrew" {
			t.Errorf("GetByID() CancelReason = %q, want %q", got.CancelReason, "supplier withdrew")
		}
		if got.BuyerStartConfirmedAt != nil {
			t.Errorf("GetByID() BuyerStartConfirmedAt = %v, want nil", got.BuyerStartConfirmedAt)
		}
	})

	t.Run("In progress transaction round trip", func(t *testing.T) {
		got, err := db.GetByID(context.Background(), testTransactionID3)
		if err != nil {
			t.Fatalf("GetByID() unexpected error: %v", err)
		}

		if got.Status != domain.TransactionInProgress {
			t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.TransactionInProgress)
		}
		if got.BuyerStartConfirmedAt == nil || !got.BuyerStartConfirmedAt.Equal(buyerStart) {
			t.Errorf("GetByID() BuyerStartConfirmedAt = %v, want %v", got.BuyerStartConfirmedAt, buyerStart)
		}
		if got.SupplierStartConfirmedAt == nil || !got.SupplierStartConfirmedAt.Equal(supplierStart) {
			t.Errorf("GetByID() SupplierStartConfirmedAt = %v, want %v", got.SupplierStartConfirmedAt, supplierStart)
		}
		if got.BuyerDeliveryConfirmedAt == nil || !got.BuyerDeliveryConfirmedAt.Equal(buyerDelivery) {
			t.Errorf("GetByID() BuyerDeliveryConfirmedAt = %v, want %v", got.BuyerDeliveryConfirmedAt, buyerDelivery)
		}
		if got.SupplierDeliveryConfirmedAt == nil || !got.SupplierDeliveryConfirmedAt.Equal(supplierDelivery) {
			t.Errorf("GetByID() SupplierDeliveryConfirmedAt = %v, want %v", got.SupplierDeliveryConfirmedAt, supplierDelivery)
		}
	})

	t.Run("Transaction Not Found", func(t *testing.T) {
		_, err := db.GetByID(context.Background(), testTransactionNotFoundID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("GetByID() error = %v, wantErr %v", err, domain.ErrNotFound)
		}
	})
}

func TestTransactionCreateDuplicateMatchRejected(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	first := newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime)
	if err := db.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	duplicate := newTransactionFixture(uuid.New(), testTransactionMatchID, fixedTime.Add(time.Minute))
	err := db.Create(context.Background(), duplicate)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Errorf("Create() duplicate error = %v, want %v", err, domain.ErrDuplicate)
	}

	transactions, err := db.List(context.Background(), testTransactionMatchID)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(transactions) != 1 {
		t.Errorf("List() got %d transactions after rejected duplicate, want 1", len(transactions))
	}
}

func TestTransactionList(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	fixtures := []*domain.Transaction{
		newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime),
		newTransactionFixture(testTransactionID2, testTransactionMatchID2, fixedTime.Add(time.Minute)),
	}

	for _, transaction := range fixtures {
		if err := db.Create(context.Background(), transaction); err != nil {
			t.Fatalf("Create() transaction %s: %v", transaction.ID, err)
		}
	}

	tests := []struct {
		Name          string
		MatchID       uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Match sees its own transaction",
			MatchID:       testTransactionMatchID,
			ExpectedIDs:   []uuid.UUID{testTransactionID},
			ExpectedTotal: 1,
		},
		{
			Name:          "Match without transactions sees none",
			MatchID:       testTransactionMatchID4,
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
		{
			Name:          "Unknown match sees none",
			MatchID:       uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.List(context.Background(), tt.MatchID)
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("List() got %d transactions, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, transaction := range got {
				found[transaction.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("List() did not return transaction %v", wantID)
				}
			}
		})
	}
}

func TestTransactionListByRequest(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	fixtures := []*domain.Transaction{
		newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime),
		newTransactionFixture(testTransactionID2, testTransactionMatchID2, fixedTime.Add(time.Minute)),
		newTransactionFixture(testTransactionID3, testTransactionMatchID3, fixedTime.Add(2*time.Minute)),
	}

	for _, transaction := range fixtures {
		if err := db.Create(context.Background(), transaction); err != nil {
			t.Fatalf("Create() transaction %s: %v", transaction.ID, err)
		}
	}

	tests := []struct {
		Name            string
		SupplyRequestID uuid.UUID
		ExpectedIDs     []uuid.UUID
		ExpectedTotal   int
	}{
		{
			Name:            "First request sees transactions of its two matches",
			SupplyRequestID: testTransactionRequestID,
			ExpectedIDs:     []uuid.UUID{testTransactionID, testTransactionID2},
			ExpectedTotal:   2,
		},
		{
			Name:            "Second request sees the transaction of its match",
			SupplyRequestID: testTransactionRequestID2,
			ExpectedIDs:     []uuid.UUID{testTransactionID3},
			ExpectedTotal:   1,
		},
		{
			Name:            "Unknown request sees no transactions",
			SupplyRequestID: uuid.New(),
			ExpectedIDs:     nil,
			ExpectedTotal:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ListByRequest(context.Background(), tt.SupplyRequestID)
			if err != nil {
				t.Fatalf("ListByRequest() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("ListByRequest() got %d transactions, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, transaction := range got {
				found[transaction.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListByRequest() did not return transaction %v", wantID)
				}
			}
		})
	}
}

func TestTransactionGetByMatch(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	saved := newTransactionFixture(testTransactionID, testTransactionMatchID2, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	tests := []struct {
		Name        string
		MatchID     uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Transaction found by match",
			MatchID:     testTransactionMatchID2,
			ExpectedErr: nil,
		},
		{
			Name:        "Match without transaction not found",
			MatchID:     testTransactionMatchID4,
			ExpectedErr: domain.ErrNotFound,
		},
		{
			Name:        "Unknown match not found",
			MatchID:     uuid.New(),
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.GetByMatch(context.Background(), tt.MatchID)

			if tt.ExpectedErr != nil {
				if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("GetByMatch() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByMatch() unexpected error: %v", err)
			}

			if got.ID != saved.ID {
				t.Errorf("GetByMatch() ID = %v, want %v", got.ID, saved.ID)
			}
			if got.MatchID != saved.MatchID {
				t.Errorf("GetByMatch() MatchID = %v, want %v", got.MatchID, saved.MatchID)
			}
			if got.Status != saved.Status {
				t.Errorf("GetByMatch() Status = %v, want %v", got.Status, saved.Status)
			}
			if !got.CreatedAt.Equal(saved.CreatedAt) {
				t.Errorf("GetByMatch() CreatedAt = %v, want %v", got.CreatedAt, saved.CreatedAt)
			}
		})
	}
}

func TestTransactionUpdate(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	created := newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated, err := db.GetByID(context.Background(), testTransactionID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	buyerConfirm := fixedTime.Add(10 * time.Minute)
	supplierConfirm := fixedTime.Add(11 * time.Minute)
	if err := updated.ConfirmStart(domain.BuyerParticipant, buyerConfirm); err != nil {
		t.Fatalf("ConfirmStart() buyer error: %v", err)
	}
	if err := updated.ConfirmStart(domain.SupplierParticipant, supplierConfirm); err != nil {
		t.Fatalf("ConfirmStart() supplier error: %v", err)
	}
	updated.UpdatedAt = fixedTime.Add(12 * time.Minute)

	if err := db.Update(context.Background(), &updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testTransactionID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}

	if got.Status != domain.TransactionInProgress {
		t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.TransactionInProgress)
	}
	if got.BuyerStartConfirmedAt == nil || !got.BuyerStartConfirmedAt.Equal(buyerConfirm) {
		t.Errorf("GetByID() BuyerStartConfirmedAt = %v, want %v", got.BuyerStartConfirmedAt, buyerConfirm)
	}
	if got.SupplierStartConfirmedAt == nil || !got.SupplierStartConfirmedAt.Equal(supplierConfirm) {
		t.Errorf("GetByID() SupplierStartConfirmedAt = %v, want %v", got.SupplierStartConfirmedAt, supplierConfirm)
	}
	if got.BuyerDeliveryConfirmedAt != nil {
		t.Errorf("GetByID() BuyerDeliveryConfirmedAt = %v, want nil", got.BuyerDeliveryConfirmedAt)
	}
	if got.CancelledBy != nil {
		t.Errorf("GetByID() CancelledBy = %v, want nil", got.CancelledBy)
	}
	if !got.UpdatedAt.Equal(fixedTime.Add(12 * time.Minute)) {
		t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, fixedTime.Add(12*time.Minute))
	}
	if got.MatchID != created.MatchID {
		t.Errorf("GetByID() MatchID = %v, want unchanged %v", got.MatchID, created.MatchID)
	}
}

func TestTransactionDelete(t *testing.T) {
	setupTransactionTestData(t)
	db := repository.NewTransactionRepository(TestPool)

	created := newTransactionFixture(testTransactionID, testTransactionMatchID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := db.Delete(context.Background(), testTransactionID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetByID(context.Background(), testTransactionID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}

	_, err = db.GetByMatch(context.Background(), testTransactionMatchID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByMatch() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}
