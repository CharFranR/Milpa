package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

// TestSavePersistsFullLocation covers the geolocation defect end to end: the
// addresses row used to be created only when Department was set, and no
// production path ever set it, so latitude and longitude were never stored.
func TestSavePersistsFullLocation(t *testing.T) {
	cleanupTables(t)
	db := repository.NewUserRepository(TestPool)
	ctx := context.Background()

	user := basicUser(testUserID)
	user.Address = domain.Address{
		Department:   "Leon",
		Municipality: "Leon",
		AddressLine:  "Costado Sur del Parque Central",
		Latitude:     12.434322610629877,
		Longitude:    -86.87891042845236,
	}

	if _, err := db.Save(ctx, user); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}

	if saved.Address.ID == uuid.Nil {
		t.Fatal("FindByID() Address.ID = nil UUID, want the persisted address row")
	}
	if saved.Address.Department != "Leon" {
		t.Errorf("department = %q, want %q", saved.Address.Department, "Leon")
	}
	if saved.Address.Municipality != "Leon" {
		t.Errorf("municipality = %q, want %q", saved.Address.Municipality, "Leon")
	}
	if saved.Address.AddressLine != "Costado Sur del Parque Central" {
		t.Errorf("address line = %q, want %q", saved.Address.AddressLine, "Costado Sur del Parque Central")
	}
	if saved.Address.Latitude != 12.434322610629877 {
		t.Errorf("latitude = %v, want 12.434322610629877", saved.Address.Latitude)
	}
	if saved.Address.Longitude != -86.87891042845236 {
		t.Errorf("longitude = %v, want -86.87891042845236", saved.Address.Longitude)
	}
	if !saved.Address.HasCoordinates() {
		t.Error("HasCoordinates() = false, want true for a registered farmer with coordinates")
	}
}

// TestSavePersistsAddressLineWithoutDepartment pins the dead gate: a farmer who
// only supplied a free-text line still owns that line.
func TestSavePersistsAddressLineWithoutDepartment(t *testing.T) {
	cleanupTables(t)
	db := repository.NewUserRepository(TestPool)
	ctx := context.Background()

	user := basicUser(testUserID)
	user.Address = domain.Address{AddressLine: "Calle Ruben Dario"}

	if _, err := db.Save(ctx, user); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if saved.Address.ID == uuid.Nil {
		t.Fatal("FindByID() Address.ID = nil UUID, want a row for the free-text address line")
	}
	if saved.Address.AddressLine != "Calle Ruben Dario" {
		t.Errorf("address line = %q, want %q", saved.Address.AddressLine, "Calle Ruben Dario")
	}
}

// TestUpdateDoesNotNullAddressID is the PATCH /users/{id} orphan: the update
// used to rebuild the address from the request, which zeroed the address id
// and wrote NULL into users.address_id.
func TestUpdateDoesNotNullAddressID(t *testing.T) {
	cleanupTables(t)
	db := repository.NewUserRepository(TestPool)
	ctx := context.Background()

	user := basicUser(testUserID)
	user.Address = domain.Address{
		Department:   "Leon",
		Municipality: "Leon",
		AddressLine:  "Costado Sur del Parque Central",
		Latitude:     12.434322610629877,
		Longitude:    -86.87891042845236,
	}
	if _, err := db.Save(ctx, user); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	addressID := saved.Address.ID

	// A partial update that says nothing about the address.
	saved.PhoneNumber = "9999-0000"
	if err := db.Update(ctx, saved); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	var storedAddressID *uuid.UUID
	if err := TestPool.QueryRow(ctx, "SELECT address_id FROM users WHERE id = $1", testUserID).Scan(&storedAddressID); err != nil {
		t.Fatalf("select address_id: %v", err)
	}
	if storedAddressID == nil {
		t.Fatal("users.address_id = NULL after a partial update, want the original address")
	}
	if *storedAddressID != addressID {
		t.Errorf("users.address_id = %v, want %v", *storedAddressID, addressID)
	}

	after, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() after partial update: %v", err)
	}
	if after.Address.Department != "Leon" {
		t.Errorf("department = %q after a partial update, want %q", after.Address.Department, "Leon")
	}
	if after.Address.Latitude != 12.434322610629877 {
		t.Errorf("latitude = %v after a partial update, want 12.434322610629877", after.Address.Latitude)
	}
}

// TestUpdateChangesCoordinates is the other half: a farmer who moves must be
// able to change the stored coordinates.
func TestUpdateChangesCoordinates(t *testing.T) {
	cleanupTables(t)
	db := repository.NewUserRepository(TestPool)
	ctx := context.Background()

	user := basicUser(testUserID)
	user.Address = domain.Address{
		Department:   "Leon",
		Municipality: "Leon",
		AddressLine:  "Costado Sur del Parque Central",
		Latitude:     12.434322610629877,
		Longitude:    -86.87891042845236,
	}
	if _, err := db.Save(ctx, user); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}

	saved.Address.AddressLine = "Barrio San Francisco"
	saved.Address.Latitude = 13.0913
	saved.Address.Longitude = -86.0014
	if err := db.Update(ctx, saved); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	after, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() after coordinate change: %v", err)
	}
	if after.Address.Latitude != 13.0913 {
		t.Errorf("latitude = %v, want 13.0913", after.Address.Latitude)
	}
	if after.Address.Longitude != -86.0014 {
		t.Errorf("longitude = %v, want -86.0014", after.Address.Longitude)
	}
	if after.Address.AddressLine != "Barrio San Francisco" {
		t.Errorf("address line = %q, want %q", after.Address.AddressLine, "Barrio San Francisco")
	}
	if after.Address.ID != saved.Address.ID {
		t.Errorf("address id = %v, want the unchanged %v", after.Address.ID, saved.Address.ID)
	}
}

// TestUpdateCreatesAddressForUserWithoutOne covers the user that registered with
// no address at all and supplies one later.
func TestUpdateCreatesAddressForUserWithoutOne(t *testing.T) {
	cleanupTables(t)
	db := repository.NewUserRepository(TestPool)
	ctx := context.Background()

	if _, err := db.Save(ctx, basicUser(testUserID)); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}
	if saved.Address.ID != uuid.Nil {
		t.Fatalf("Address.ID = %v, want nil for a user registered without an address", saved.Address.ID)
	}

	saved.Address = domain.Address{
		Department:   "Jinotega",
		Municipality: "Jinotega",
		AddressLine:  "Barrio Centro",
		Latitude:     13.0913,
		Longitude:    -86.0014,
	}
	if err := db.Update(ctx, saved); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	after, err := db.FindByID(ctx, testUserID)
	if err != nil {
		t.Fatalf("FindByID() after adding an address: %v", err)
	}
	if after.Address.ID == uuid.Nil {
		t.Fatal("Address.ID = nil UUID, want the address row created by the update")
	}
	if after.Address.Department != "Jinotega" || after.Address.Latitude != 13.0913 {
		t.Errorf("address = %+v, want the Jinotega address with latitude 13.0913", after.Address)
	}
}
