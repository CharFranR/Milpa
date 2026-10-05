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

var testInventorySupplierID uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e501")
var testInventoryOtherSupplierID uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e502")

var testInventoryID uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e511")
var testInventoryID2 uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e512")
var testInventoryID3 uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e513")
var testInventoryNotFoundID uuid.UUID = uuid.MustParse("e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e519")

func newSupplierInventoryFixture(id, supplierID uuid.UUID, productName string, quantity float64, createdAt time.Time) *domain.SupplierInventory {
	return &domain.SupplierInventory{
		ID:          id,
		SupplierID:  supplierID,
		ProductName: productName,
		Quantity:    quantity,
		AmountUnit:  domain.Kg,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}

func setupSupplierInventoryTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testInventorySupplierID,
			FirstName:    "Inventory",
			LastName:     "Supplier",
			Role:         domain.RoleAgricultor,
			Email:        "inventory-supplier@example.com",
			PhoneNumber:  "4500-0001",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testInventoryOtherSupplierID,
			FirstName:    "Inventory",
			LastName:     "OtherSupplier",
			Role:         domain.RoleAgricultor,
			Email:        "inventory-other-supplier@example.com",
			PhoneNumber:  "4500-0002",
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

func TestSupplierInventoryCreateAndGetByID(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	saved := newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if saved.ID != testInventoryID {
		t.Errorf("Create() mutated ID = %v, want %v", saved.ID, testInventoryID)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testInventoryID,
			ExpectedErr: nil,
		},
		{
			Name:        "Inventory Not Found",
			ID:          testInventoryNotFoundID,
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
			if got.ProductName != saved.ProductName {
				t.Errorf("GetByID() ProductName = %v, want %v", got.ProductName, saved.ProductName)
			}
			if got.Quantity != saved.Quantity {
				t.Errorf("GetByID() Quantity = %v, want %v", got.Quantity, saved.Quantity)
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

func TestSupplierInventoryCreateDuplicateRejected(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	first := newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime)
	if err := db.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	duplicate := newSupplierInventoryFixture(uuid.New(), testInventorySupplierID, "Maize", 20, fixedTime.Add(time.Minute))
	err := db.Create(context.Background(), duplicate)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Errorf("Create() duplicate error = %v, want %v", err, domain.ErrDuplicate)
	}

	entries, err := db.ListBySupplier(context.Background(), testInventorySupplierID)
	if err != nil {
		t.Fatalf("ListBySupplier() error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("ListBySupplier() got %d entries after rejected duplicate, want 1", len(entries))
	}

	otherSupplier := newSupplierInventoryFixture(uuid.New(), testInventoryOtherSupplierID, "Maize", 30, fixedTime.Add(time.Minute))
	if err := db.Create(context.Background(), otherSupplier); err != nil {
		t.Errorf("Create() for another supplier with same product error: %v", err)
	}
}

func TestSupplierInventoryListBySupplier(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	fixtures := []*domain.SupplierInventory{
		newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime),
		newSupplierInventoryFixture(testInventoryID2, testInventorySupplierID, "Beans", 40, fixedTime.Add(time.Minute)),
		newSupplierInventoryFixture(testInventoryID3, testInventoryOtherSupplierID, "Maize", 88, fixedTime.Add(2*time.Minute)),
	}

	for _, inventory := range fixtures {
		if err := db.Create(context.Background(), inventory); err != nil {
			t.Fatalf("Create() inventory %s: %v", inventory.ID, err)
		}
	}

	tests := []struct {
		Name          string
		SupplierID    uuid.UUID
		ExpectedIDs   []uuid.UUID
		ExpectedTotal int
	}{
		{
			Name:          "Supplier sees only their own inventory",
			SupplierID:    testInventorySupplierID,
			ExpectedIDs:   []uuid.UUID{testInventoryID, testInventoryID2},
			ExpectedTotal: 2,
		},
		{
			Name:          "Other supplier sees only their own inventory",
			SupplierID:    testInventoryOtherSupplierID,
			ExpectedIDs:   []uuid.UUID{testInventoryID3},
			ExpectedTotal: 1,
		},
		{
			Name:          "Supplier without inventory sees none",
			SupplierID:    uuid.New(),
			ExpectedIDs:   nil,
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.ListBySupplier(context.Background(), tt.SupplierID)
			if err != nil {
				t.Fatalf("ListBySupplier() unexpected error: %v", err)
			}

			if len(got) != tt.ExpectedTotal {
				t.Fatalf("ListBySupplier() got %d entries, want %d", len(got), tt.ExpectedTotal)
			}

			found := make(map[uuid.UUID]bool, len(got))
			for _, inventory := range got {
				found[inventory.ID] = true
			}
			for _, wantID := range tt.ExpectedIDs {
				if !found[wantID] {
					t.Errorf("ListBySupplier() did not return inventory %v", wantID)
				}
			}
		})
	}
}

func TestSupplierInventoryFindBySupplierAndProduct(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	saved := newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime)
	if err := db.Create(context.Background(), saved); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	other := newSupplierInventoryFixture(testInventoryID3, testInventoryOtherSupplierID, "Beans", 88, fixedTime.Add(time.Minute))
	if err := db.Create(context.Background(), other); err != nil {
		t.Fatalf("Create() other error: %v", err)
	}

	tests := []struct {
		Name        string
		SupplierID  uuid.UUID
		ProductName string
		ExpectedErr error
	}{
		{
			Name:        "Inventory found for supplier and product",
			SupplierID:  testInventorySupplierID,
			ProductName: "Maize",
			ExpectedErr: nil,
		},
		{
			Name:        "Not found for unknown product",
			SupplierID:  testInventorySupplierID,
			ProductName: "Rice",
			ExpectedErr: domain.ErrNotFound,
		},
		{
			Name:        "Not found for another supplier",
			SupplierID:  testInventoryOtherSupplierID,
			ProductName: "Maize",
			ExpectedErr: domain.ErrNotFound,
		},
		{
			Name:        "Not found for unknown supplier",
			SupplierID:  uuid.New(),
			ProductName: "Maize",
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := db.FindBySupplierAndProduct(context.Background(), tt.SupplierID, tt.ProductName)

			if tt.ExpectedErr != nil {
				if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindBySupplierAndProduct() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("FindBySupplierAndProduct() unexpected error: %v", err)
			}

			if got.ID != saved.ID {
				t.Errorf("FindBySupplierAndProduct() ID = %v, want %v", got.ID, saved.ID)
			}
			if got.SupplierID != saved.SupplierID {
				t.Errorf("FindBySupplierAndProduct() SupplierID = %v, want %v", got.SupplierID, saved.SupplierID)
			}
			if got.ProductName != saved.ProductName {
				t.Errorf("FindBySupplierAndProduct() ProductName = %v, want %v", got.ProductName, saved.ProductName)
			}
			if got.Quantity != saved.Quantity {
				t.Errorf("FindBySupplierAndProduct() Quantity = %v, want %v", got.Quantity, saved.Quantity)
			}
			if got.AmountUnit != saved.AmountUnit {
				t.Errorf("FindBySupplierAndProduct() AmountUnit = %v, want %v", got.AmountUnit, saved.AmountUnit)
			}
		})
	}
}

func TestSupplierInventoryUpdate(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	created := newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated, err := db.GetByID(context.Background(), testInventoryID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	updated.ProductName = "Maize Premium"
	updated.Quantity = 99.5
	updated.AmountUnit = domain.Tn
	updated.UpdatedAt = fixedTime.Add(5 * time.Minute)

	if err := db.Update(context.Background(), &updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := db.GetByID(context.Background(), testInventoryID)
	if err != nil {
		t.Fatalf("GetByID() after Update error: %v", err)
	}

	if got.ProductName != "Maize Premium" {
		t.Errorf("GetByID() ProductName = %q, want %q", got.ProductName, "Maize Premium")
	}
	if got.Quantity != 99.5 {
		t.Errorf("GetByID() Quantity = %v, want %v", got.Quantity, 99.5)
	}
	if got.AmountUnit != domain.Tn {
		t.Errorf("GetByID() AmountUnit = %v, want %v", got.AmountUnit, domain.Tn)
	}
	if !got.UpdatedAt.Equal(fixedTime.Add(5 * time.Minute)) {
		t.Errorf("GetByID() UpdatedAt = %v, want %v", got.UpdatedAt, fixedTime.Add(5*time.Minute))
	}
	if got.SupplierID != created.SupplierID {
		t.Errorf("GetByID() SupplierID = %v, want unchanged %v", got.SupplierID, created.SupplierID)
	}
	if got.CreatedAt.Equal(got.UpdatedAt) {
		t.Error("GetByID() CreatedAt equals UpdatedAt, want CreatedAt preserved")
	}
}

func TestSupplierInventoryDelete(t *testing.T) {
	setupSupplierInventoryTestData(t)
	db := repository.NewSupplierInventoryRepository(TestPool)

	created := newSupplierInventoryFixture(testInventoryID, testInventorySupplierID, "Maize", 12.75, fixedTime)
	if err := db.Create(context.Background(), created); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := db.Delete(context.Background(), testInventoryID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := db.GetByID(context.Background(), testInventoryID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}

	_, err = db.FindBySupplierAndProduct(context.Background(), testInventorySupplierID, "Maize")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindBySupplierAndProduct() after Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}
