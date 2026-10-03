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

var testSupplyOfferSupplierID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b201")
var testSupplyOfferOtherSupplierID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b202")
var testSupplyOfferBuyerID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b203")

var testSupplyOfferRequestID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b211")
var testSupplyOfferRequestID2 uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b212")

var testSupplyOfferID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b221")
var testSupplyOfferID2 uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b222")
var testSupplyOfferID3 uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b223")
var testSupplyOfferNotFoundID uuid.UUID = uuid.MustParse("b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b229")

func newSupplyOfferFixture(id, supplierID, supplyRequestID uuid.UUID, createdAt time.Time) *domain.SupplyOffer {
	return &domain.SupplyOffer{
		ID:                  id,
		SupplierID:          supplierID,
		SupplyRequest:       supplyRequestID,
		TotalAmount:         420.5,
		AmountUnit:          domain.Tn,
		ProposedDeliveryDay: fixedTime.Add(48 * time.Hour),
		DeliveryAvailable:   true,
		Status:              domain.OfferActive,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}
}

func newSupplyOfferRequestFixture(id, buyerID uuid.UUID, productName string) *domain.SupplyRequest {
	return &domain.SupplyRequest{
		ID:               id,
		BuyerID:          buyerID,
		ProductName:      productName,
		TotalAmount:      100,
		ActualAmount:     100,
		AmountUnit:       domain.Kg,
		NumberOfUnits:    10,
		AmountPerUnit:    10,
		UnitOfMeasure:    domain.Kg,
		RequestDeadline:  fixedTime.Add(24 * time.Hour),
		DeliveryDeadline: fixedTime.Add(48 * time.Hour),
		Description:      "seed request",
		Status:           domain.SupplyRequestOpen,
		CreatedAt:        fixedTime,
		UpdatedAt:        fixedTime,
	}
}

func setupSupplyOfferTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)
	requestRepo := repository.NewSupplyRequestRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testSupplyOfferSupplierID,
			FirstName:    "Supply",
			LastName:     "OfferSupplier",
			Role:         domain.RoleAgricultor,
			Email:        "supply-offer-supplier@example.com",
			PhoneNumber:  "4200-0001",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testSupplyOfferOtherSupplierID,
			FirstName:    "Supply",
			LastName:     "OfferOther",
			Role:         domain.RoleAgricultor,
			Email:        "supply-offer-other@example.com",
			PhoneNumber:  "4200-0002",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testSupplyOfferBuyerID,
			FirstName:    "Supply",
			LastName:     "OfferBuyer",
			Role:         domain.RoleCompradorMinorista,
			Email:        "supply-offer-buyer@example.com",
			PhoneNumber:  "4200-0003",
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
		newSupplyOfferRequestFixture(testSupplyOfferRequestID, testSupplyOfferBuyerID, "Maize"),
		newSupplyOfferRequestFixture(testSupplyOfferRequestID2, testSupplyOfferBuyerID, "Beans"),
	}

	for _, supplyRequest := range requests {
		if err := requestRepo.Create(context.Background(), supplyRequest); err != nil {
			t.Fatalf("insert fixture supply request %s: %v", supplyRequest.ID, err)
		}
	}
}

func TestSupplyOfferCreateAndGetByID(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	saved := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if saved.ID != testSupplyOfferID {
		t.Errorf("Create() mutated ID = %v, want %v", saved.ID, testSupplyOfferID)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testSupplyOfferID,
			ExpectedErr: nil,
		},
		{
			Name:        "Supply Offer Not Found",
			ID:          testSupplyOfferNotFoundID,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.GetByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("GetByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByID() unexpected error: %v", err)
			}

			if got.ID != saved.ID {
				t.Errorf("GetByID() ID = %v, want %v", got.ID, saved.ID)
			}
			if got.SupplierID != saved.SupplierID {
				t.Errorf("GetByID() SupplierID = %v, want %v", got.SupplierID, saved.SupplierID)
			}
			if got.SupplyRequest != saved.SupplyRequest {
				t.Errorf("GetByID() SupplyRequest = %v, want %v", got.SupplyRequest, saved.SupplyRequest)
			}
			if got.TotalAmount != saved.TotalAmount {
				t.Errorf("GetByID() TotalAmount = %v, want %v", got.TotalAmount, saved.TotalAmount)
			}
			if got.AmountUnit != saved.AmountUnit {
				t.Errorf("GetByID() AmountUnit = %v, want %v", got.AmountUnit, saved.AmountUnit)
			}
			if !got.ProposedDeliveryDay.Equal(saved.ProposedDeliveryDay) {
				t.Errorf("GetByID() ProposedDeliveryDay = %v, want %v", got.ProposedDeliveryDay, saved.ProposedDeliveryDay)
			}
			if got.DeliveryAvailable != saved.DeliveryAvailable {
				t.Errorf("GetByID() DeliveryAvailable = %v, want %v", got.DeliveryAvailable, saved.DeliveryAvailable)
			}
			if got.Status != saved.Status {
				t.Errorf("GetByID() Status = %v, want %v", got.Status, saved.Status)
			}
			if !got.CreatedAt.Equal(saved.CreatedAt) {
				t.Errorf("GetByID() CreatedAt = %v, want %v", got.CreatedAt, saved.CreatedAt)
			}
			if !got.UpdatedAt.Equal(saved.UpdatedAt) {
				t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, saved.UpdatedAt)
			}
		})
	}
}

func TestSupplyOfferCreateDuplicateRejected(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	first := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	if err := db.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	duplicate := newSupplyOfferFixture(uuid.New(), testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime.Add(time.Minute))
	err := db.Create(context.Background(), duplicate)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Errorf("Create() duplicate error = %v, want %v", err, domain.ErrDuplicate)
	}

	offers, err := db.ListByRequest(context.Background(), testSupplyOfferRequestID)
	if err != nil {
		t.Fatalf("ListByRequest() error: %v", err)
	}
	if len(offers) != 1 {
		t.Errorf("ListByRequest() got %d offers after rejected duplicate, want 1", len(offers))
	}
}

func TestSupplyOfferList(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	fixtures := []*domain.SupplyOffer{
		newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime),
		newSupplyOfferFixture(testSupplyOfferID2, testSupplyOfferOtherSupplierID, testSupplyOfferRequestID, fixedTime.Add(time.Minute)),
		newSupplyOfferFixture(testSupplyOfferID3, testSupplyOfferSupplierID, testSupplyOfferRequestID2, fixedTime.Add(2*time.Minute)),
	}

	for _, offer := range fixtures {
		if err := db.Create(context.Background(), offer); err != nil {
			t.Fatalf("Create() offer %s: %v", offer.ID, err)
		}
	}

	tests := []struct {
		Name          string
		SupplierID    uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Supplier sees only their own offers",
			SupplierID:    testSupplyOfferSupplierID,
			ExpectedIDs:   []uuid.UUID{testSupplyOfferID, testSupplyOfferID3},
			ExpectedTotal: 2,
		},
		{
			Name:          "Other supplier sees only their own offer",
			SupplierID:    testSupplyOfferOtherSupplierID,
			ExpectedIDs:   []uuid.UUID{testSupplyOfferID2},
			ExpectedTotal: 1,
		},
		{
			Name:          "Supplier without offers sees none",
			SupplierID:    uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.List(context.Background(), tt.SupplierID)
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("List() got %d offers, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, offer := range got {
				found[offer.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("List() did not return offer %v", wantID)
				}
			}
		})
	}
}

func TestSupplyOfferListByRequest(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	fixtures := []*domain.SupplyOffer{
		newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime),
		newSupplyOfferFixture(testSupplyOfferID2, testSupplyOfferOtherSupplierID, testSupplyOfferRequestID, fixedTime.Add(time.Minute)),
		newSupplyOfferFixture(testSupplyOfferID3, testSupplyOfferSupplierID, testSupplyOfferRequestID2, fixedTime.Add(2*time.Minute)),
	}

	for _, offer := range fixtures {
		if err := db.Create(context.Background(), offer); err != nil {
			t.Fatalf("Create() offer %s: %v", offer.ID, err)
		}
	}

	tests := []struct {
		Name            string
		SupplyRequestID uuid.UUID
		ExpectedIDs     []uuid.UUID
		ExpectedTotal   int
	}{
		{
			Name:            "Request sees offers from both suppliers",
			SupplyRequestID: testSupplyOfferRequestID,
			ExpectedIDs:     []uuid.UUID{testSupplyOfferID, testSupplyOfferID2},
			ExpectedTotal:   2,
		},
		{
			Name:            "Second request sees only its own offer",
			SupplyRequestID: testSupplyOfferRequestID2,
			ExpectedIDs:     []uuid.UUID{testSupplyOfferID3},
			ExpectedTotal:   1,
		},
		{
			Name:            "Unknown request sees no offers",
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
				t.Fatalf("ListByRequest() got %d offers, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, offer := range got {
				found[offer.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListByRequest() did not return offer %v", wantID)
				}
			}
		})
	}
}

func TestSupplyOfferFindBySupplierAndRequest(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	saved := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	other := newSupplyOfferFixture(testSupplyOfferID2, testSupplyOfferOtherSupplierID, testSupplyOfferRequestID2, fixedTime.Add(time.Minute))
	if err := db.Create(context.Background(), other); err != nil {
		t.Fatalf("Create() other offer error: %v", err)
	}

	tests := []struct {
		Name            string
		SupplierID      uuid.UUID
		SupplyRequestID uuid.UUID
		ExpectedErr     error
	}{
		{
			Name:            "Offer found for supplier and request",
			SupplierID:      testSupplyOfferSupplierID,
			SupplyRequestID: testSupplyOfferRequestID,
			ExpectedErr:     nil,
		},
		{
			Name:            "Not found for another supplier on the same request",
			SupplierID:      testSupplyOfferOtherSupplierID,
			SupplyRequestID: testSupplyOfferRequestID,
			ExpectedErr:     domain.ErrNotFound,
		},
		{
			Name:            "Not found for unknown request",
			SupplierID:      testSupplyOfferSupplierID,
			SupplyRequestID: uuid.New(),
			ExpectedErr:     domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.FindBySupplierAndRequest(context.Background(), tt.SupplierID, tt.SupplyRequestID)

			if tt.ExpectedErr != nil {
				if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindBySupplierAndRequest() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("FindBySupplierAndRequest() unexpected error: %v", err)
			}

			if got.ID != saved.ID {
				t.Errorf("FindBySupplierAndRequest() ID = %v, want %v", got.ID, saved.ID)
			}
			if got.SupplierID != saved.SupplierID {
				t.Errorf("FindBySupplierAndRequest() SupplierID = %v, want %v", got.SupplierID, saved.SupplierID)
			}
			if got.SupplyRequest != saved.SupplyRequest {
				t.Errorf("FindBySupplierAndRequest() SupplyRequest = %v, want %v", got.SupplyRequest, saved.SupplyRequest)
			}
			if got.TotalAmount != saved.TotalAmount {
				t.Errorf("FindBySupplierAndRequest() TotalAmount = %v, want %v", got.TotalAmount, saved.TotalAmount)
			}
			if got.Status != saved.Status {
				t.Errorf("FindBySupplierAndRequest() Status = %v, want %v", got.Status, saved.Status)
			}
			if !got.CreatedAt.Equal(saved.CreatedAt) {
				t.Errorf("FindBySupplierAndRequest() CreatedAt = %v, want %v", got.CreatedAt, saved.CreatedAt)
			}
		})
	}
}

func TestSupplyOfferUpdate(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	created := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated, err := db.GetByID(context.Background(), testSupplyOfferID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if err := updated.MarkMatched(); err != nil {
		t.Fatalf("MarkMatched() error: %v", err)
	}
	updated.TotalAmount = 500.25
	updated.DeliveryAvailable = false
	updated.UpdatedAt = fixedTime.Add(5 * time.Minute)

	if err := db.Update(context.Background(), &updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testSupplyOfferID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}

	if got.Status != domain.OfferMatched {
		t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.OfferMatched)
	}
	if got.TotalAmount != 500.25 {
		t.Errorf("GetByID() TotalAmount = %v, want %v", got.TotalAmount, 500.25)
	}
	if got.DeliveryAvailable {
		t.Error("GetByID() DeliveryAvailable = true, want false")
	}
	if !got.UpdatedAt.Equal(fixedTime.Add(5 * time.Minute)) {
		t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, fixedTime.Add(5*time.Minute))
	}
	if got.SupplierID != created.SupplierID {
		t.Errorf("GetByID() SupplierID = %v, want unchanged %v", got.SupplierID, created.SupplierID)
	}
	if got.SupplyRequest != created.SupplyRequest {
		t.Errorf("GetByID() SupplyRequest = %v, want unchanged %v", got.SupplyRequest, created.SupplyRequest)
	}
}

func TestSupplyOfferPriceAndCommentsRoundTrip(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)
	ctx := context.Background()

	price, repriced := 12.75, 9.5

	saved := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	saved.PricePerUnit = &price
	saved.Comments = "Picked yesterday, delivered chilled"
	if err := db.Create(ctx, saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := db.GetByID(ctx, testSupplyOfferID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if got.PricePerUnit == nil || *got.PricePerUnit != price {
		t.Errorf("GetByID() PricePerUnit = %v, want %v", got.PricePerUnit, price)
	}
	if got.Comments != saved.Comments {
		t.Errorf("GetByID() Comments = %q, want %q", got.Comments, saved.Comments)
	}

	// The locked read spells its own column list, so pin it to the unlocked one.
	// A price missing from the FOR UPDATE projection would hand the transactional
	// paths a different value for the same row, silently: each read stays
	// self-consistent on its own.
	locked, err := db.LockByIDForUpdate(ctx, testSupplyOfferID)
	if err != nil {
		t.Fatalf("LockByIDForUpdate() error: %v", err)
	}
	if locked.PricePerUnit == nil || *locked.PricePerUnit != price {
		t.Errorf("LockByIDForUpdate() PricePerUnit = %v, want %v", locked.PricePerUnit, price)
	}
	if locked.Comments != got.Comments {
		t.Errorf("LockByIDForUpdate() Comments = %q, want %q", locked.Comments, got.Comments)
	}

	got.PricePerUnit = &repriced
	got.Comments = "Bulk price if you take the lot"
	if err := db.Update(ctx, &got); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	reread, err := db.GetByID(ctx, testSupplyOfferID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}
	if reread.PricePerUnit == nil || *reread.PricePerUnit != repriced {
		t.Errorf("GetByID() after Update PricePerUnit = %v, want %v", reread.PricePerUnit, repriced)
	}
	if reread.Comments != got.Comments {
		t.Errorf("GetByID() after Update Comments = %q, want %q", reread.Comments, got.Comments)
	}
}

// TestSupplyOfferPriceCheckIsTheStorageBackstop proves the column is nullable
// and that the CHECK is what keeps it honest. The use case is the layer that
// refuses an unpriced create; this is the layer that still holds for a direct
// write, and the two are not redundant.
func TestSupplyOfferPriceCheckIsTheStorageBackstop(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)
	ctx := context.Background()

	t.Run("nil price is accepted so legacy offers survive", func(t *testing.T) {
		legacy := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
		if err := db.Create(ctx, legacy); err != nil {
			t.Fatalf("Create() with a nil price error: %v", err)
		}
		got, err := db.GetByID(ctx, testSupplyOfferID)
		if err != nil {
			t.Fatalf("GetByID() error: %v", err)
		}
		if got.PricePerUnit != nil {
			t.Errorf("GetByID() PricePerUnit = %v, want nil", *got.PricePerUnit)
		}
	})

	t.Run("a zero price is rejected by the CHECK", func(t *testing.T) {
		zero := 0.0
		rejected := newSupplyOfferFixture(testSupplyOfferID2, testSupplyOfferOtherSupplierID, testSupplyOfferRequestID, fixedTime.Add(time.Minute))
		rejected.PricePerUnit = &zero
		if err := db.Create(ctx, rejected); err == nil {
			t.Fatal("Create() with a zero price succeeded, want ck_supply_offers_price to reject it")
		}
	})
}

func TestSupplyOfferDelete(t *testing.T) {
	setupSupplyOfferTestData(t)
	db := repository.NewSupplyOfferRepository(TestPool)

	created := newSupplyOfferFixture(testSupplyOfferID, testSupplyOfferSupplierID, testSupplyOfferRequestID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := db.Delete(context.Background(), testSupplyOfferID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetByID(context.Background(), testSupplyOfferID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}

	offers, err := db.ListByRequest(context.Background(), testSupplyOfferRequestID)
	if err != nil {
		t.Fatalf("ListByRequest() after Delete error: %v", err)
	}
	if len(offers) != 0 {
		t.Errorf("ListByRequest() after Delete got %d offers, want 0", len(offers))
	}
}
