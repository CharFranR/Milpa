package integration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

func catalogueOffering(ownerID uuid.UUID, name string, createdAt time.Time) *domain.Offering {
	offering, err := domain.NewOffering(ownerID, name, domain.OfferingProduct, createdAt)
	if err != nil {
		panic(err)
	}
	return offering
}

func offeringNames(offerings []domain.Offering) map[string]bool {
	names := make(map[string]bool, len(offerings))
	for _, offering := range offerings {
		names[offering.Name] = true
	}
	return names
}

func TestOfferingCatalogueReadHidesExpiredAndDeactivated(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	fresh := catalogueOffering(testOwnerID, "Maiz fresco", now)
	expired := catalogueOffering(testOwnerID, "Maiz viejo", now)
	expiredExpiry := now.Add(-24 * time.Hour)
	expired.ExpiresAt = &expiredExpiry

	retired := catalogueOffering(testOwnerID, "Maiz retirado", now)
	retired.Deactivate()

	stillGood := catalogueOffering(testOwnerID, "Frijol fresco", now)

	for _, offering := range []*domain.Offering{fresh, expired, retired, stillGood} {
		if err := repo.Save(ctx, offering); err != nil {
			t.Fatalf("save %s: %v", offering.Name, err)
		}
	}

	listed, err := repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() error: %v", err)
	}

	names := offeringNames(listed)
	if names["Maiz viejo"] {
		t.Error("the catalogue returned an expired product")
	}
	if names["Maiz retirado"] {
		t.Error("the catalogue returned a deactivated product")
	}
	if !names["Maiz fresco"] || !names["Frijol fresco"] {
		t.Errorf("the catalogue dropped live products: %v", names)
	}

	expired.Renew(now.Add(72*time.Hour), now)
	if err := repo.Update(ctx, expired); err != nil {
		t.Fatalf("renew: %v", err)
	}

	listed, err = repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() after the renewal error: %v", err)
	}
	if !offeringNames(listed)["Maiz viejo"] {
		t.Error("Renew() did not bring the product back to the catalogue")
	}
	if !expired.IsActive {
		t.Error("Renew() left the product inactive")
	}
	if expired.UpdatedAt != now {
		t.Errorf("Renew() updated_at = %v, want %v", expired.UpdatedAt, now)
	}
}

func TestOfferingFindByIDStillReadsAHiddenOffering(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	offering := catalogueOffering(testOwnerID, "Maiz viejo", now)
	offering.Deactivate()
	expiry := now.Add(-time.Hour)
	offering.ExpiresAt = &expiry

	if err := repo.Save(ctx, offering); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, offering.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if loaded.ID != offering.ID {
		t.Errorf("FindByID() = %v, want %v", loaded.ID, offering.ID)
	}
	if loaded.ExpiresAt == nil || !loaded.ExpiresAt.Equal(expiry) {
		t.Errorf("FindByID() expires_at = %v, want %v", loaded.ExpiresAt, expiry)
	}
	if loaded.IsActive {
		t.Error("FindByID() reports a deactivated offering as active")
	}
}

func TestOfferingSaveRefusesADuplicateActiveProduct(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	createdAt := time.Now().UTC()

	first := catalogueOffering(testOwnerID, "Maiz criollo", createdAt)
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}

	second := catalogueOffering(testOwnerID, "Maiz criollo", createdAt)
	if err := repo.Save(ctx, second); !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("save duplicate = %v, want ErrDuplicate", err)
	}

	listed, err := repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() error: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("the catalogue holds %d products, want only the first one", len(listed))
	}
}

func TestOfferingSaveRefusesADuplicateActiveProductPublishedLater(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	first := catalogueOffering(testOwnerID, "Maiz criollo", now)
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}

	second := catalogueOffering(testOwnerID, "Maiz criollo", now.Add(3*time.Hour))
	err := repo.Save(ctx, second)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("save duplicate published three hours later = %v, want ErrDuplicate", err)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		t.Errorf("the driver error leaked past the repository: %v", pgErr)
	}

	listed, err := repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() error: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("the catalogue holds %d products, want only the first one", len(listed))
	}
	if listed[0].ID != first.ID {
		t.Errorf("the catalogue kept %s, want the first one %s", listed[0].ID, first.ID)
	}
}

func TestOfferingSaveAcceptsTheSameNameOnceTheFirstIsDeactivated(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	first := catalogueOffering(testOwnerID, "Frijol rojo", now)
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}

	first.Deactivate()
	if err := repo.Update(ctx, first); err != nil {
		t.Fatalf("deactivate first: %v", err)
	}

	second := catalogueOffering(testOwnerID, "Frijol rojo", now.Add(4*time.Hour))
	if err := repo.Save(ctx, second); err != nil {
		t.Fatalf("save after deactivating the first: %v", err)
	}

	listed, err := repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("the catalogue holds %d products, want only the republished one", len(listed))
	}
	if listed[0].ID != second.ID {
		t.Errorf("the catalogue kept %s, want the republished %s", listed[0].ID, second.ID)
	}
}

func TestOfferingDuplicatePreventionIndexCoversOnlyTheActivePair(t *testing.T) {
	setupOfferingTestData(t)

	ctx := context.Background()

	rows, err := TestPool.Query(ctx,
		`SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'offerings' AND indexname LIKE 'uq_offerings_farmer%'`)
	if err != nil {
		t.Fatalf("read the offerings unique indexes: %v", err)
	}
	defer rows.Close()

	definitions := map[string]string{}
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatalf("scan an index definition: %v", err)
		}
		definitions[name] = definition
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read the index definitions: %v", err)
	}

	if _, stale := definitions["uq_offerings_farmer_product_window"]; stale {
		t.Error("uq_offerings_farmer_product_window is still installed")
	}

	active, ok := definitions["uq_offerings_farmer_active_product"]
	if !ok {
		t.Fatalf("uq_offerings_farmer_active_product is missing; installed: %v", definitions)
	}
	if strings.Contains(active, "created_at") {
		t.Errorf("index definition = %q, want a key without created_at", active)
	}
	if !strings.Contains(active, "WHERE") || !strings.Contains(active, "is_active") {
		t.Errorf("index definition = %q, want a partial index over the active rows", active)
	}
}

func TestOfferingSaveRefusesANegativeQuantity(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)

	offering := catalogueOffering(testOwnerID, "Maiz", time.Now().UTC())
	offering.QuantityAvailable = -1

	if err := repo.Save(context.Background(), offering); err == nil {
		t.Fatal("Save() of a negative quantity = nil, want the storage check to refuse it")
	}
}

func TestOfferingStoreKeepsTheProductLocation(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	unitID := seededUnitOfMeasure(t, "kg")
	categoryID := uuid.New()

	offering := catalogueOffering(testOwnerID, "Maiz del cerro", time.Now().UTC())
	offering.Variety = "Cuzqueño"
	offering.UnitOfMeasureID = &unitID
	offering.QuantityAvailable = 250
	offering.CategoryID = &categoryID
	latitude := 11.9747
	longitude := -86.0941
	offering.Latitude = &latitude
	offering.Longitude = &longitude

	if _, err := TestPool.Exec(ctx,
		`INSERT INTO categories (id, name, main_category, is_active) VALUES ($1, 'Granos de prueba', 'granos', TRUE)`,
		categoryID); err != nil {
		t.Fatalf("insert category: %v", err)
	}

	if err := repo.Save(ctx, offering); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, offering.ID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}

	if loaded.Variety != "Cuzqueño" {
		t.Errorf("variety = %q, want Cuzqueño", loaded.Variety)
	}
	if loaded.UnitOfMeasureID == nil || *loaded.UnitOfMeasureID != unitID {
		t.Errorf("unit of measure = %v, want %v", loaded.UnitOfMeasureID, unitID)
	}
	if loaded.QuantityAvailable != 250 {
		t.Errorf("quantity = %v, want 250", loaded.QuantityAvailable)
	}
	if loaded.CategoryID == nil || *loaded.CategoryID != categoryID {
		t.Errorf("category = %v, want %v", loaded.CategoryID, categoryID)
	}
	if loaded.Latitude == nil || *loaded.Latitude != latitude {
		t.Errorf("latitude = %v, want %v", loaded.Latitude, latitude)
	}
	if loaded.Longitude == nil || *loaded.Longitude != longitude {
		t.Errorf("longitude = %v, want %v", loaded.Longitude, longitude)
	}
}

func TestOfferingStoreRefusesCoordinatesOffTheGlobe(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)

	offering := catalogueOffering(testOwnerID, "Maiz", time.Now().UTC())
	latitude := 120.0
	longitude := -86.0
	offering.Latitude = &latitude
	offering.Longitude = &longitude

	if err := repo.Save(context.Background(), offering); err == nil {
		t.Fatal("Save() with a latitude above 90 = nil, want the storage check to refuse it")
	}
}

func TestOfferingGrandfatheredRowIsHiddenUntilItIsCompleted(t *testing.T) {
	setupOfferingTestData(t)

	repo := repository.NewOfferingRepository(TestPool)
	ctx := context.Background()

	legacy := &domain.Offering{
		ID:        uuid.New(),
		UserID:    testOwnerID,
		Type:      domain.OfferingProduct,
		Name:      "Producto heredado",
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, legacy); err != nil {
		t.Fatalf("save legacy row: %v", err)
	}

	if err := legacy.RequirePublishable(); !errors.Is(err, domain.ErrVarietyRequired) {
		t.Errorf("a row written before the catalogue columns existed passes the gate: %v", err)
	}

	unitID := seededUnitOfMeasure(t, "unidad")
	categoryID := uuid.New()
	if _, err := TestPool.Exec(ctx,
		`INSERT INTO categories (id, name, main_category, is_active) VALUES ($1, 'Hortalizas de prueba', 'hortalizas', TRUE)`,
		categoryID); err != nil {
		t.Fatalf("insert category: %v", err)
	}

	legacy.Variety = "Cebollín"
	legacy.UnitOfMeasureID = &unitID
	legacy.QuantityAvailable = 40
	legacy.CategoryID = &categoryID
	if err := legacy.RequirePublishable(); err != nil {
		t.Fatalf("a completed row still fails the gate: %v", err)
	}

	if err := repo.Update(ctx, legacy); err != nil {
		t.Fatalf("update legacy row: %v", err)
	}

	listed, err := repo.FindByUserID(ctx, testOwnerID)
	if err != nil {
		t.Fatalf("FindByUserID() error: %v", err)
	}
	if !offeringNames(listed)["Producto heredado"] {
		t.Error("the completed product is still missing from the catalogue")
	}
}
