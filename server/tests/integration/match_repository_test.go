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

var testMatchSupplierID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c301")
var testMatchOtherSupplierID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c302")
var testMatchBuyerID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c303")

var testMatchRequestID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c311")
var testMatchRequestID2 uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c312")

var testMatchOfferID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c321")
var testMatchOfferID2 uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c322")
var testMatchOfferID3 uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c323")

var testMatchID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c331")
var testMatchID2 uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c332")
var testMatchID3 uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c333")
var testMatchNotFoundID uuid.UUID = uuid.MustParse("c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c339")

func newMatchFixture(id, offerID, requestID uuid.UUID, createdAt time.Time) *domain.Match {
	return &domain.Match{
		ID:            id,
		SupplyOffer:   offerID,
		SupplyRequest: requestID,
		Status:        domain.MatchActive,
		MatchedAmount: 120.75,
		AmountUnit:    domain.Kg,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
}

func setupMatchTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)
	requestRepo := repository.NewSupplyRequestRepository(TestPool)
	offerRepo := repository.NewSupplyOfferRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testMatchSupplierID,
			FirstName:    "Match",
			LastName:     "Supplier",
			Role:         domain.RoleAgricultor,
			Email:        "match-supplier@example.com",
			PhoneNumber:  "4300-0001",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testMatchOtherSupplierID,
			FirstName:    "Match",
			LastName:     "OtherSupplier",
			Role:         domain.RoleAgricultor,
			Email:        "match-other-supplier@example.com",
			PhoneNumber:  "4300-0002",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testMatchBuyerID,
			FirstName:    "Match",
			LastName:     "Buyer",
			Role:         domain.RoleCompradorMinorista,
			Email:        "match-buyer@example.com",
			PhoneNumber:  "4300-0003",
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
		newSupplyOfferRequestFixture(testMatchRequestID, testMatchBuyerID, "Maize"),
		newSupplyOfferRequestFixture(testMatchRequestID2, testMatchBuyerID, "Beans"),
	}

	for _, supplyRequest := range requests {
		if err := requestRepo.Create(context.Background(), supplyRequest); err != nil {
			t.Fatalf("insert fixture supply request %s: %v", supplyRequest.ID, err)
		}
	}

	offers := []*domain.SupplyOffer{
		newSupplyOfferFixture(testMatchOfferID, testMatchSupplierID, testMatchRequestID, fixedTime),
		newSupplyOfferFixture(testMatchOfferID2, testMatchOtherSupplierID, testMatchRequestID, fixedTime.Add(time.Minute)),
		newSupplyOfferFixture(testMatchOfferID3, testMatchSupplierID, testMatchRequestID2, fixedTime.Add(2*time.Minute)),
	}

	for _, offer := range offers {
		if err := offerRepo.Create(context.Background(), offer); err != nil {
			t.Fatalf("insert fixture supply offer %s: %v", offer.ID, err)
		}
	}
}

func seedMatchFixtures(t *testing.T, db *repository.MatchRepositoryImpl) {
	t.Helper()

	fixtures := []*domain.Match{
		newMatchFixture(testMatchID, testMatchOfferID, testMatchRequestID, fixedTime),
		newMatchFixture(testMatchID2, testMatchOfferID2, testMatchRequestID, fixedTime.Add(time.Minute)),
		newMatchFixture(testMatchID3, testMatchOfferID3, testMatchRequestID2, fixedTime.Add(2*time.Minute)),
	}

	for _, match := range fixtures {
		if err := db.Create(context.Background(), match); err != nil {
			t.Fatalf("Create() match %s: %v", match.ID, err)
		}
	}
}

func cancelMatch(t *testing.T, db *repository.MatchRepositoryImpl, id uuid.UUID) {
	t.Helper()

	match, err := db.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if err := match.Cancel(); err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}
	match.UpdatedAt = fixedTime.Add(10 * time.Minute)
	if err := db.Update(context.Background(), match); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
}

func TestMatchCreateAndGetByID(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)

	saved := newMatchFixture(testMatchID, testMatchOfferID, testMatchRequestID, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if saved.ID != testMatchID {
		t.Errorf("Create() mutated ID = %v, want %v", saved.ID, testMatchID)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testMatchID,
			ExpectedErr: nil,
		},
		{
			Name:        "Match Not Found",
			ID:          testMatchNotFoundID,
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.GetByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if got != nil {
					t.Errorf("GetByID() got = %v, want nil", got)
				}
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
			if got.SupplyOffer != saved.SupplyOffer {
				t.Errorf("GetByID() SupplyOffer = %v, want %v", got.SupplyOffer, saved.SupplyOffer)
			}
			if got.SupplyRequest != saved.SupplyRequest {
				t.Errorf("GetByID() SupplyRequest = %v, want %v", got.SupplyRequest, saved.SupplyRequest)
			}
			if got.Status != saved.Status {
				t.Errorf("GetByID() Status = %v, want %v", got.Status, saved.Status)
			}
			if got.MatchedAmount != saved.MatchedAmount {
				t.Errorf("GetByID() MatchedAmount = %v, want %v", got.MatchedAmount, saved.MatchedAmount)
			}
			if got.AmountUnit != saved.AmountUnit {
				t.Errorf("GetByID() AmountUnit = %v, want %v", got.AmountUnit, saved.AmountUnit)
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

func TestMatchCreateDuplicateOfferRejected(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)

	first := newMatchFixture(testMatchID, testMatchOfferID, testMatchRequestID, fixedTime)
	if err := db.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	duplicate := newMatchFixture(uuid.New(), testMatchOfferID, testMatchRequestID, fixedTime.Add(time.Minute))
	err := db.Create(context.Background(), duplicate)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Errorf("Create() duplicate error = %v, want %v", err, domain.ErrDuplicate)
	}

	matches, err := db.ListByOffer(context.Background(), testMatchOfferID)
	if err != nil {
		t.Fatalf("ListByOffer() error: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("ListByOffer() got %d matches after rejected duplicate, want 1", len(matches))
	}
}

func TestMatchListByOffer(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)

	tests := []struct {
		Name          string
		OfferID       uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "First offer has its own match",
			OfferID:       testMatchOfferID,
			ExpectedIDs:   []uuid.UUID{testMatchID},
			ExpectedTotal: 1,
		},
		{
			Name:          "Third offer has its own match",
			OfferID:       testMatchOfferID3,
			ExpectedIDs:   []uuid.UUID{testMatchID3},
			ExpectedTotal: 1,
		},
		{
			Name:          "Unknown offer has no matches",
			OfferID:       uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ListByOffer(context.Background(), tt.OfferID)
			if err != nil {
				t.Fatalf("ListByOffer() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("ListByOffer() got %d matches, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, match := range got {
				found[match.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListByOffer() did not return match %v", wantID)
				}
			}
		})
	}
}

func TestMatchListByRequest(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)

	tests := []struct {
		Name            string
		SupplyRequestID uuid.UUID
		ExpectedIDs     []uuid.UUID
		ExpectedTotal   int
	}{
		{
			Name:            "First request has two matches",
			SupplyRequestID: testMatchRequestID,
			ExpectedIDs:     []uuid.UUID{testMatchID, testMatchID2},
			ExpectedTotal:   2,
		},
		{
			Name:            "Second request has one match",
			SupplyRequestID: testMatchRequestID2,
			ExpectedIDs:     []uuid.UUID{testMatchID3},
			ExpectedTotal:   1,
		},
		{
			Name:            "Unknown request has no matches",
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
				t.Fatalf("ListByRequest() got %d matches, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, match := range got {
				found[match.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListByRequest() did not return match %v", wantID)
				}
			}
		})
	}
}

func TestMatchListActiveByRequest(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)
	cancelMatch(t, db, testMatchID3)

	tests := []struct {
		Name            string
		SupplyRequestID uuid.UUID
		ExpectedIDs     []uuid.UUID
		ExpectedTotal   int
	}{
		{
			Name:            "First request has two active matches",
			SupplyRequestID: testMatchRequestID,
			ExpectedIDs:     []uuid.UUID{testMatchID, testMatchID2},
			ExpectedTotal:   2,
		},
		{
			Name:            "Second request has no active matches after cancel",
			SupplyRequestID: testMatchRequestID2,
			ExpectedIDs:     nil,
			ExpectedTotal:   0,
		},
		{
			Name:            "Unknown request has no active matches",
			SupplyRequestID: uuid.New(),
			ExpectedIDs:     nil,
			ExpectedTotal:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ListActiveByRequest(context.Background(), tt.SupplyRequestID)
			if err != nil {
				t.Fatalf("ListActiveByRequest() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("ListActiveByRequest() got %d matches, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, match := range got {
				found[match.ID] = true
				if match.Status != domain.MatchActive {
					t.Errorf("ListActiveByRequest() returned match %v with status %v", match.ID, match.Status)
				}
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListActiveByRequest() did not return match %v", wantID)
				}
			}
		})
	}
}

func TestMatchListActiveBySupplier(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)
	cancelMatch(t, db, testMatchID3)

	tests := []struct {
		Name          string
		SupplierID    uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Supplier sees only their own active matches",
			SupplierID:    testMatchSupplierID,
			ExpectedIDs:   []uuid.UUID{testMatchID},
			ExpectedTotal: 1,
		},
		{
			Name:          "Other supplier sees only their own active match",
			SupplierID:    testMatchOtherSupplierID,
			ExpectedIDs:   []uuid.UUID{testMatchID2},
			ExpectedTotal: 1,
		},
		{
			Name:          "Supplier without matches sees none",
			SupplierID:    uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ListActiveBySupplier(context.Background(), tt.SupplierID)
			if err != nil {
				t.Fatalf("ListActiveBySupplier() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("ListActiveBySupplier() got %d matches, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, match := range got {
				found[match.ID] = true
				if match.Status != domain.MatchActive {
					t.Errorf("ListActiveBySupplier() returned match %v with status %v", match.ID, match.Status)
				}
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListActiveBySupplier() did not return match %v", wantID)
				}
			}
		})
	}
}

func TestMatchExistsActiveByRequest(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)
	cancelMatch(t, db, testMatchID3)

	tests := []struct {
		Name            string
		SupplyRequestID uuid.UUID
		Expected        bool
	}{
		{
			Name:            "Request with active matches exists",
			SupplyRequestID: testMatchRequestID,
			Expected:        true,
		},
		{
			Name:            "Request with only cancelled matches does not exist",
			SupplyRequestID: testMatchRequestID2,
			Expected:        false,
		},
		{
			Name:            "Unknown request does not exist",
			SupplyRequestID: uuid.New(),
			Expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ExistsActiveByRequest(context.Background(), tt.SupplyRequestID)
			if err != nil {
				t.Fatalf("ExistsActiveByRequest() unexpected error: %v", err)
			}
			if got != tt.Expected {
				t.Errorf("ExistsActiveByRequest() = %v, want %v", got, tt.Expected)
			}
		})
	}
}

func TestMatchExistsActiveByOffer(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)
	seedMatchFixtures(t, db)
	cancelMatch(t, db, testMatchID3)

	tests := []struct {
		Name     string
		OfferID  uuid.UUID
		Expected bool
	}{
		{
			Name:     "Offer with active match exists",
			OfferID:  testMatchOfferID,
			Expected: true,
		},
		{
			Name:     "Offer with only cancelled match does not exist",
			OfferID:  testMatchOfferID3,
			Expected: false,
		},
		{
			Name:     "Unknown offer does not exist",
			OfferID:  uuid.New(),
			Expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ExistsActiveByOffer(context.Background(), tt.OfferID)
			if err != nil {
				t.Fatalf("ExistsActiveByOffer() unexpected error: %v", err)
			}
			if got != tt.Expected {
				t.Errorf("ExistsActiveByOffer() = %v, want %v", got, tt.Expected)
			}
		})
	}
}

func TestMatchUpdate(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)

	created := newMatchFixture(testMatchID, testMatchOfferID, testMatchRequestID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated, err := db.GetByID(context.Background(), testMatchID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if err := updated.Cancel(); err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}
	updated.MatchedAmount = 60.5
	updated.AmountUnit = domain.Tn
	updated.UpdatedAt = fixedTime.Add(15 * time.Minute)

	if err := db.Update(context.Background(), updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testMatchID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}

	if got.Status != domain.MatchCancelled {
		t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.MatchCancelled)
	}
	if got.MatchedAmount != 60.5 {
		t.Errorf("GetByID() MatchedAmount = %v, want %v", got.MatchedAmount, 60.5)
	}
	if got.AmountUnit != domain.Tn {
		t.Errorf("GetByID() AmountUnit = %v, want %v", got.AmountUnit, domain.Tn)
	}
	if !got.UpdatedAt.Equal(fixedTime.Add(15 * time.Minute)) {
		t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, fixedTime.Add(15*time.Minute))
	}
	if got.SupplyOffer != created.SupplyOffer {
		t.Errorf("GetByID() SupplyOffer = %v, want unchanged %v", got.SupplyOffer, created.SupplyOffer)
	}
	if got.SupplyRequest != created.SupplyRequest {
		t.Errorf("GetByID() SupplyRequest = %v, want unchanged %v", got.SupplyRequest, created.SupplyRequest)
	}

	exists, err := db.ExistsActiveByOffer(context.Background(), testMatchOfferID)
	if err != nil {
		t.Fatalf("ExistsActiveByOffer() error: %v", err)
	}
	if exists {
		t.Error("ExistsActiveByOffer() = true after cancel, want false")
	}
}

func TestMatchDelete(t *testing.T) {
	setupMatchTestData(t)
	db := repository.NewMatchRepository(TestPool)

	created := newMatchFixture(testMatchID, testMatchOfferID, testMatchRequestID, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := db.Delete(context.Background(), testMatchID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testMatchID)
	if got != nil {
		t.Errorf("GetByID() after Delete got = %v, want nil", got)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}
