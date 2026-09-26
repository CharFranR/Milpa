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

var testSupplyRequestBuyerID uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a101")
var testSupplyRequestOtherBuyerID uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a102")
var testSupplyRequestID uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a111")
var testSupplyRequestID2 uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a112")
var testSupplyRequestID3 uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a113")
var testSupplyRequestNotFoundID uuid.UUID = uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a199")

func newSupplyRequestFixture(id, buyerID uuid.UUID, productName string, createdAt time.Time) *domain.SupplyRequest {
	return &domain.SupplyRequest{
		ID:            id,
		BuyerID:       buyerID,
		ProductName:   productName,
		TotalAmount:   150.5,
		ActualAmount:  150.5,
		AmountUnit:    domain.Kg,
		NumberOfUnits: 10,
		AmountPerUnit: 15.05,
		UnitOfMeasure: domain.Lb,
		Address: domain.Address{
			Department:   "Leon",
			Municipality: "Nueva Rosita",
			AddressLine:  "Calle Central 123",
			Latitude:     12.5,
			Longitude:    -86.25,
		},
		RequestDeadline:      fixedTime.Add(24 * time.Hour),
		DeliveryDeadline:     fixedTime.Add(72 * time.Hour),
		Description:          "Fresh maize for the month",
		MultipleProviders:    true,
		MinAmountPerProvider: 25.5,
		Status:               domain.SupplyRequestOpen,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	}
}

func setupSupplyRequestTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testSupplyRequestBuyerID,
			FirstName:    "Supply",
			LastName:     "RequestBuyer",
			Role:         domain.RoleMIPYME,
			Email:        "supply-request-buyer@example.com",
			PhoneNumber:  "4100-0001",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testSupplyRequestOtherBuyerID,
			FirstName:    "Supply",
			LastName:     "RequestOther",
			Role:         domain.RoleMIPYME,
			Email:        "supply-request-other@example.com",
			PhoneNumber:  "4100-0002",
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
}

func TestSupplyRequestCreateAndGetByID(t *testing.T) {
	setupSupplyRequestTestData(t)
	db := repository.NewSupplyRequestRepository(TestPool)

	saved := newSupplyRequestFixture(testSupplyRequestID, testSupplyRequestBuyerID, "Maize", fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if saved.ID != testSupplyRequestID {
		t.Errorf("Create() mutated ID = %v, want %v", saved.ID, testSupplyRequestID)
	}

	noAddress := newSupplyRequestFixture(testSupplyRequestID2, testSupplyRequestBuyerID, "Beans", fixedTime.Add(time.Minute))
	noAddress.Address = domain.Address{}
	if err := db.Create(context.Background(), noAddress); err != nil {
		t.Fatalf("Create() without address error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path with address",
			ID:          testSupplyRequestID,
			ExpectedErr: nil,
		},
		{
			Name:        "Supply Request Not Found",
			ID:          testSupplyRequestNotFoundID,
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
			if got.BuyerID != saved.BuyerID {
				t.Errorf("GetByID() BuyerID = %v, want %v", got.BuyerID, saved.BuyerID)
			}
			if got.ProductName != saved.ProductName {
				t.Errorf("GetByID() ProductName = %v, want %v", got.ProductName, saved.ProductName)
			}
			if got.TotalAmount != saved.TotalAmount {
				t.Errorf("GetByID() TotalAmount = %v, want %v", got.TotalAmount, saved.TotalAmount)
			}
			if got.ActualAmount != saved.ActualAmount {
				t.Errorf("GetByID() ActualAmount = %v, want %v", got.ActualAmount, saved.ActualAmount)
			}
			if got.AmountUnit != saved.AmountUnit {
				t.Errorf("GetByID() AmountUnit = %v, want %v", got.AmountUnit, saved.AmountUnit)
			}
			if got.NumberOfUnits != saved.NumberOfUnits {
				t.Errorf("GetByID() NumberOfUnits = %v, want %v", got.NumberOfUnits, saved.NumberOfUnits)
			}
			if got.AmountPerUnit != saved.AmountPerUnit {
				t.Errorf("GetByID() AmountPerUnit = %v, want %v", got.AmountPerUnit, saved.AmountPerUnit)
			}
			if got.UnitOfMeasure != saved.UnitOfMeasure {
				t.Errorf("GetByID() UnitOfMeasure = %v, want %v", got.UnitOfMeasure, saved.UnitOfMeasure)
			}
			if got.Address.ID == uuid.Nil {
				t.Error("GetByID() Address.ID is nil, want inserted address id")
			}
			if got.Address.Department != saved.Address.Department {
				t.Errorf("GetByID() Address.Department = %v, want %v", got.Address.Department, saved.Address.Department)
			}
			if got.Address.Municipality != saved.Address.Municipality {
				t.Errorf("GetByID() Address.Municipality = %v, want %v", got.Address.Municipality, saved.Address.Municipality)
			}
			if got.Address.AddressLine != saved.Address.AddressLine {
				t.Errorf("GetByID() Address.AddressLine = %v, want %v", got.Address.AddressLine, saved.Address.AddressLine)
			}
			if got.Address.Latitude != saved.Address.Latitude {
				t.Errorf("GetByID() Address.Latitude = %v, want %v", got.Address.Latitude, saved.Address.Latitude)
			}
			if got.Address.Longitude != saved.Address.Longitude {
				t.Errorf("GetByID() Address.Longitude = %v, want %v", got.Address.Longitude, saved.Address.Longitude)
			}
			if !got.RequestDeadline.Equal(saved.RequestDeadline) {
				t.Errorf("GetByID() RequestDeadline = %v, want %v", got.RequestDeadline, saved.RequestDeadline)
			}
			if !got.DeliveryDeadline.Equal(saved.DeliveryDeadline) {
				t.Errorf("GetByID() DeliveryDeadline = %v, want %v", got.DeliveryDeadline, saved.DeliveryDeadline)
			}
			if got.Description != saved.Description {
				t.Errorf("GetByID() Description = %v, want %v", got.Description, saved.Description)
			}
			if got.MultipleProviders != saved.MultipleProviders {
				t.Errorf("GetByID() MultipleProviders = %v, want %v", got.MultipleProviders, saved.MultipleProviders)
			}
			if got.MinAmountPerProvider != saved.MinAmountPerProvider {
				t.Errorf("GetByID() MinAmountPerProvider = %v, want %v", got.MinAmountPerProvider, saved.MinAmountPerProvider)
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

	t.Run("Created without address reads back with zero address", func(t *testing.T) {
		got, err := db.GetByID(context.Background(), testSupplyRequestID2)
		if err != nil {
			t.Fatalf("GetByID() unexpected error: %v", err)
		}
		if got.Address.ID != uuid.Nil {
			t.Errorf("GetByID() Address.ID = %v, want nil UUID", got.Address.ID)
		}
		if got.Address.Department != "" {
			t.Errorf("GetByID() Address.Department = %q, want empty", got.Address.Department)
		}
		if got.Address.AddressLine != "" {
			t.Errorf("GetByID() Address.AddressLine = %q, want empty", got.Address.AddressLine)
		}
	})
}

func TestSupplyRequestList(t *testing.T) {
	setupSupplyRequestTestData(t)
	db := repository.NewSupplyRequestRepository(TestPool)

	fixtures := []*domain.SupplyRequest{
		newSupplyRequestFixture(testSupplyRequestID, testSupplyRequestBuyerID, "Maize", fixedTime),
		newSupplyRequestFixture(testSupplyRequestID2, testSupplyRequestBuyerID, "Beans", fixedTime.Add(time.Minute)),
		newSupplyRequestFixture(testSupplyRequestID3, testSupplyRequestOtherBuyerID, "Rice", fixedTime),
	}

	for _, supplyRequest := range fixtures {
		if err := db.Create(context.Background(), supplyRequest); err != nil {
			t.Fatalf("Create() supply request %s: %v", supplyRequest.ID, err)
		}
	}

	tests := []struct {
		Name          string
		BuyerID       uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Buyer sees only their own supply requests",
			BuyerID:       testSupplyRequestBuyerID,
			ExpectedIDs:   []uuid.UUID{testSupplyRequestID, testSupplyRequestID2},
			ExpectedTotal: 2,
		},
		{
			Name:          "Other buyer sees only their own supply request",
			BuyerID:       testSupplyRequestOtherBuyerID,
			ExpectedIDs:   []uuid.UUID{testSupplyRequestID3},
			ExpectedTotal: 1,
		},
		{
			Name:          "Buyer without supply requests sees none",
			BuyerID:       uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.List(context.Background(), tt.BuyerID)
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("List() got %d supply requests, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, supplyRequest := range got {
				found[supplyRequest.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("List() did not return supply request %v", wantID)
				}
			}
		})
	}
}

func TestSupplyRequestUpdate(t *testing.T) {
	setupSupplyRequestTestData(t)
	db := repository.NewSupplyRequestRepository(TestPool)

	created := newSupplyRequestFixture(testSupplyRequestID, testSupplyRequestBuyerID, "Maize", fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated, err := db.GetByID(context.Background(), testSupplyRequestID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	updated.Description = "Updated description"
	updated.ActualAmount = 100.25
	updated.MultipleProviders = false
	updated.MinAmountPerProvider = 99.75
	if err := updated.Cancel(); err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}
	updated.UpdatedAt = fixedTime.Add(10 * time.Minute)

	if err := db.Update(context.Background(), &updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testSupplyRequestID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}

	if got.Description != "Updated description" {
		t.Errorf("GetByID() Description = %q, want %q", got.Description, "Updated description")
	}
	if got.ActualAmount != 100.25 {
		t.Errorf("GetByID() ActualAmount = %v, want %v", got.ActualAmount, 100.25)
	}
	if got.MultipleProviders {
		t.Error("GetByID() MultipleProviders = true, want false")
	}
	if got.MinAmountPerProvider != 99.75 {
		t.Errorf("GetByID() MinAmountPerProvider = %v, want %v", got.MinAmountPerProvider, 99.75)
	}
	if got.Status != domain.SupplyRequestCancelled {
		t.Errorf("GetByID() Status = %v, want %v", got.Status, domain.SupplyRequestCancelled)
	}
	if !got.UpdatedAt.Equal(fixedTime.Add(10 * time.Minute)) {
		t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, fixedTime.Add(10*time.Minute))
	}
	if got.TotalAmount != created.TotalAmount {
		t.Errorf("GetByID() TotalAmount = %v, want unchanged %v", got.TotalAmount, created.TotalAmount)
	}
	if got.Address.ID == uuid.Nil {
		t.Error("GetByID() Address.ID is nil, want address preserved")
	}
	if got.Address.Department != created.Address.Department {
		t.Errorf("GetByID() Address.Department = %v, want %v", got.Address.Department, created.Address.Department)
	}
}

func TestSupplyRequestDelete(t *testing.T) {
	setupSupplyRequestTestData(t)
	db := repository.NewSupplyRequestRepository(TestPool)

	created := newSupplyRequestFixture(testSupplyRequestID, testSupplyRequestBuyerID, "Maize", fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := db.Delete(context.Background(), testSupplyRequestID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetByID(context.Background(), testSupplyRequestID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}
