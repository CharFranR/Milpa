package repository_test

import (
	"context"

	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
)

var fixedTime time.Time = time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
var fixedTime2 time.Time = time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)
var testUserID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testAddressID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111112")
var nullID *uuid.UUID = (*uuid.UUID)(nil)

func TestUserSave(t *testing.T) {

	u := &domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	u_with_address := &domain.User{
		ID:        testUserID,
		FirstName: "John",
		LastName:  "Doe",
		Role:      domain.RoleOptions(1),
		Address: domain.Address{
			ID:           testAddressID,
			Department:   "Leon",
			Municipality: "Leon",
			AddressLine:  "Calle Rubén Darìo, Av. Central Nte., León",
			Latitude:     12.435010881390852,
			Longitude:    -86.87811141017944,
		},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	u_address_line_only := &domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{AddressLine: "Calle Ruben Dario"},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		user    *domain.User
		expect  func(mock pgxmock.PgxPoolIface)
	}{
		{
			name:    "happy path without address",
			wantErr: false,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {

				m.ExpectBegin()

				rows := pgxmock.NewRows([]string{"id"}).AddRow(testAddressID.String())

				m.ExpectQuery(`INSERT INTO users`).
					WithArgs(u.ID, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt, nullID,
						u.Email, u.PhoneNumber, u.PasswordHash, u.SuspendedAt, u.PhotoURL).WillReturnRows(rows)

				m.ExpectCommit()
			},
		},
		{
			name:    "Happy path with address",
			wantErr: false,
			user:    u_with_address,
			expect: func(m pgxmock.PgxPoolIface) {

				rows := pgxmock.NewRows([]string{"id"}).AddRow(testAddressID.String())
				rows2 := pgxmock.NewRows([]string{"id"}).AddRow(u_with_address.ID.String())

				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), u_with_address.Address.Department, u_with_address.Address.Municipality, u_with_address.Address.AddressLine, u_with_address.Address.Latitude, u_with_address.Address.Longitude).
					WillReturnRows(rows)

				m.ExpectQuery(`INSERT INTO users`).
					WithArgs(u_with_address.ID, u_with_address.FirstName, u_with_address.LastName, u_with_address.Role, u_with_address.CreatedAt, u_with_address.UpdatedAt, &testAddressID, u_with_address.Email, u_with_address.PhoneNumber, u_with_address.PasswordHash, u_with_address.SuspendedAt, u_with_address.PhotoURL).
					WillReturnRows(rows2)

				m.ExpectCommit()

			},
		},
		{
			// An address that only carries a free-text line used to be dropped
			// because the insert was gated on Department being set, so no
			// addresses row was ever created at registration.
			name:    "address line without department still inserts",
			wantErr: false,
			user:    u_address_line_only,
			expect: func(m pgxmock.PgxPoolIface) {
				addressRows := pgxmock.NewRows([]string{"id"}).AddRow(testAddressID)
				userRows := pgxmock.NewRows([]string{"id"}).AddRow(u_address_line_only.ID.String())

				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), "", "", "Calle Ruben Dario", 0.0, 0.0).
					WillReturnRows(addressRows)
				m.ExpectQuery(`INSERT INTO users`).
					WithArgs(u_address_line_only.ID, u_address_line_only.FirstName, u_address_line_only.LastName, u_address_line_only.Role, u_address_line_only.CreatedAt, u_address_line_only.UpdatedAt, &testAddressID,
						u_address_line_only.Email, u_address_line_only.PhoneNumber, u_address_line_only.PasswordHash, u_address_line_only.SuspendedAt, u_address_line_only.PhotoURL).
					WillReturnRows(userRows)
				m.ExpectCommit()
			},
		},
		{
			name:    "address insert fails",
			user:    u_with_address,
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").WithArgs(pgxmock.AnyArg(), u_with_address.Address.Department, u_with_address.Address.Municipality, u_with_address.Address.AddressLine, u_with_address.Address.Latitude, u_with_address.Address.Longitude).WillReturnError(errors.New("DB address failed"))

				m.ExpectRollback()
			},
		},
		{
			name:    "db write fails",
			wantErr: true,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`INSERT INTO users`).
					WithArgs(u.ID, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt, nullID,
						u.Email, u.PhoneNumber, u.PasswordHash, u.SuspendedAt, u.PhotoURL).
					WillReturnError(errors.New("db write failed"))

				m.ExpectRollback()
			},
		},
		{
			name:    "begin fail",
			wantErr: true,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("db begin failed"))
			},
		},
		{
			name:    "commit fail",
			wantErr: true,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(testAddressID.String())

				m.ExpectBegin()

				m.ExpectQuery(`INSERT INTO users`).
					WithArgs(u.ID, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt, nullID,
						u.Email, u.PhoneNumber, u.PasswordHash, u.SuspendedAt, u.PhotoURL).WillReturnRows(rows)

				m.ExpectCommit().WillReturnError(errors.New("db commit failed"))

			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mockPool, err := pgxmock.NewPool()
			repo := repository.NewUserRepository(mockPool)

			if err != nil {
				t.Fatalf("fail to create mock: %v", err)
			}

			defer mockPool.Close()

			tt.expect(mockPool)

			_, err = repo.Save(
				context.Background(),
				tt.user,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

		})
	}
}

func TestUserUpdate(t *testing.T) {

	user := &domain.User{
		ID:           testUserID,
		FirstName:    "Jane",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "milpa2@milpa.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "estasieslamejorcontrasenia123",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime2,
	}

	u_with_address := &domain.User{
		ID:        testUserID,
		FirstName: "Jane",
		LastName:  "Doe",
		Role:      domain.RoleOptions(1),
		Address: domain.Address{
			ID:           testAddressID,
			Department:   "Leon",
			Municipality: "Leon",
			AddressLine:  "Costado Sur del Parque Central, León",
			Latitude:     12.434322610629877,
			Longitude:    -86.87891042845236,
		},
		Email:        "jane@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	u_address_line_only := &domain.User{
		ID:           testUserID,
		FirstName:    "Jane",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{AddressLine: "Calle Ruben Dario"},
		Email:        "jane@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		user    *domain.User
		expect  func(pgxmock.PgxPoolIface)
	}{
		{
			name:    "Happy path",
			wantErr: false,
			user:    user,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(user.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec("UPDATE users").WithArgs(user.FirstName, user.LastName, user.Role, user.UpdatedAt, nullID, user.Email, user.PhoneNumber, user.PasswordHash, user.SuspendedAt, user.PhotoURL, user.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()

			},
		},
		{
			name:    "Happy path with address",
			wantErr: false,
			user:    u_with_address,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(u_with_address.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec("UPDATE addresses").WithArgs(u_with_address.Address.Department, u_with_address.Address.Municipality, u_with_address.Address.AddressLine, u_with_address.Address.Latitude, u_with_address.Address.Longitude, u_with_address.Address.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("UPDATE users").WithArgs(u_with_address.FirstName, u_with_address.LastName, u_with_address.Role, u_with_address.UpdatedAt, &testAddressID, u_with_address.Email, u_with_address.PhoneNumber, u_with_address.PasswordHash, u_with_address.SuspendedAt, u_with_address.PhotoURL, u_with_address.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			// The single PATCH /users/{id} un-referencing defect: no address id
			// and no address data must not null the existing address_id.
			name:    "no address data leaves address_id for the database to keep",
			wantErr: false,
			user:    user,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(user.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec("UPDATE users").WithArgs(user.FirstName, user.LastName, user.Role, user.UpdatedAt, nullID, user.Email, user.PhoneNumber, user.PasswordHash, user.SuspendedAt, user.PhotoURL, user.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "address data without an id inserts and re-points the user",
			wantErr: false,
			user:    u_address_line_only,
			expect: func(m pgxmock.PgxPoolIface) {
				addressRows := pgxmock.NewRows([]string{"id"}).AddRow(testAddressID)

				m.ExpectQuery("SELECT EXISTS").WithArgs(u_address_line_only.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), "", "", "Calle Ruben Dario", 0.0, 0.0).
					WillReturnRows(addressRows)
				m.ExpectExec("UPDATE users").WithArgs(u_address_line_only.FirstName, u_address_line_only.LastName, u_address_line_only.Role, u_address_line_only.UpdatedAt, &testAddressID, u_address_line_only.Email, u_address_line_only.PhoneNumber, u_address_line_only.PasswordHash, u_address_line_only.SuspendedAt, u_address_line_only.PhotoURL, u_address_line_only.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "Begin fail",
			wantErr: true,
			user:    user,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(user.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin().WillReturnError(errors.New("begin failed"))
			},
		},
		{
			name:    "update user fail",
			wantErr: true,
			user:    user,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(user.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()

				m.ExpectExec("UPDATE users").WithArgs(user.FirstName, user.LastName, user.Role, user.UpdatedAt, nullID, user.Email, user.PhoneNumber, user.PasswordHash, user.SuspendedAt, user.PhotoURL, user.ID).WillReturnError(errors.New("update user failed"))

				m.ExpectRollback()

			},
		},
		{
			name:    "update address fail",
			wantErr: true,
			user:    u_with_address,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(u_with_address.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()

				m.ExpectExec("UPDATE addresses").WithArgs(u_with_address.Address.Department, u_with_address.Address.Municipality, u_with_address.Address.AddressLine, u_with_address.Address.Latitude, u_with_address.Address.Longitude, u_with_address.Address.ID).WillReturnError(errors.New("update address failed"))

				m.ExpectRollback()
			},
		},
		{
			name:    "commit fail",
			wantErr: true,
			user:    user,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT EXISTS").WithArgs(user.ID.String()).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec("UPDATE users").WithArgs(user.FirstName, user.LastName, user.Role, user.UpdatedAt, nullID, user.Email, user.PhoneNumber, user.PasswordHash, user.SuspendedAt, user.PhotoURL, user.ID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit().WillReturnError(errors.New("db commit failed"))
				m.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPool, err := pgxmock.NewPool()

			repo := repository.NewUserRepository(mockPool)

			if err != nil {
				t.Fatalf("fail to create mock: %v", err)
			}

			defer mockPool.Close()

			tt.expect(mockPool)

			err = repo.Update(
				context.Background(),
				tt.user,
			)

			if tt.wantErr != (err != nil) {
				t.Errorf("%v no error was expected but got: %v", tt.name, err)
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %s there were unfulfilled expectations: %s", tt.name, err)
			}

		})
	}
}

// userRow builds the projection userColumns selects: the user columns plus the
// five address columns the LEFT JOIN contributes.
func userRow(u *domain.User, addressID any, a domain.Address) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id", "first_name", "last_name", "role", "created_at", "updated_at", "address_id",
		"email", "phone_number", "password_hash", "suspended_at", "photo_url",
		"department", "municipality", "address_line", "latitude", "longitude",
	}).AddRow(
		u.ID, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt, addressID,
		u.Email, u.PhoneNumber, u.PasswordHash, nil, u.PhotoURL,
		a.Department, a.Municipality, a.AddressLine, a.Latitude, a.Longitude,
	)
}

func TestFindByID(t *testing.T) {

	u := &domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	located := &domain.User{
		ID:    testUserID,
		Email: "john@example.com",
		Address: domain.Address{
			ID:           testAddressID,
			Department:   "Leon",
			Municipality: "Leon",
			AddressLine:  "Calle Ruben Dario",
			Latitude:     12.435,
			Longitude:    -86.878,
		},
	}

	tests := []struct {
		name    string
		wantErr bool
		user    *domain.User
		got     *domain.User
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name:    "Happy path",
			wantErr: false,
			user:    u,
			got:     u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT").WithArgs(u.ID).WillReturnRows(userRow(u, nil, domain.Address{}))
			},
		},
		{
			name:    "returns the joined address",
			wantErr: false,
			user:    u,
			got:     located,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT").WithArgs(u.ID).
					WillReturnRows(userRow(located, testAddressID, located.Address))
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT").WithArgs(u.ID).WillReturnError(errors.New("select failed"))
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mockPool, err := pgxmock.NewPool()

			repo := repository.NewUserRepository(mockPool)

			defer mockPool.Close()

			if err != nil {
				t.Fatalf("mockPool inicialization failed")
			}

			tt.expect(mockPool)

			got, err := repo.FindByID(
				context.Background(),
				tt.user.ID,
			)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("No error expected in %v but go %v", tt.name, err)
			}

			if err == nil && tt.got != nil {
				if got.Address != tt.got.Address {
					t.Errorf("FindByID() address = %+v, want %+v", got.Address, tt.got.Address)
				}
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectections were unfulfilled: %v", tt.name, err)
			}

		})

	}
}

func TestFindByEmail(t *testing.T) {

	u := &domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		user    *domain.User
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name:    "Happy path",
			wantErr: false,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT").WithArgs(u.Email).WillReturnRows(userRow(u, nil, domain.Address{}))
			},
		},
		{
			name:    "returns the joined address",
			wantErr: false,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				located := domain.Address{
					ID:           testAddressID,
					Department:   "Leon",
					Municipality: "Leon",
					AddressLine:  "Calle Ruben Dario",
					Latitude:     12.435,
					Longitude:    -86.878,
				}
				m.ExpectQuery("SELECT").WithArgs(u.Email).
					WillReturnRows(userRow(u, testAddressID, located))
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			user:    u,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("SELECT").WithArgs(u.Email).WillReturnError(errors.New("select failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPool, err := pgxmock.NewPool()

			if err != nil {
				t.Fatalf("mockPool init failed: %v", err)
			}

			defer mockPool.Close()

			repo := repository.NewUserRepository(mockPool)

			tt.expect(mockPool)

			_, err = repo.FindByEmail(context.Background(), tt.user.Email)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("No error expected in %v but got %v", tt.name, err)
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestExistsByEmail(t *testing.T) {

	u := &domain.User{
		ID:           testUserID,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         domain.RoleOptions(1),
		Address:      domain.Address{},
		Email:        "john@example.com",
		PhoneNumber:  "1234-5678",
		PasswordHash: "lamejorcontrasenia1233",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	tests := []struct {
		name    string
		exists  bool
		wantErr bool
		user    *domain.User
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "exists", exists: true, wantErr: false, user: u,
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)
				m.ExpectQuery("SELECT").WithArgs(u.Email).WillReturnRows(rows)
			},
		},
		{
			name: "does not exist", exists: false, wantErr: false, user: u,
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"exists"}).AddRow(false)
				m.ExpectQuery("SELECT").WithArgs(u.Email).WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("mockPool init failed: %v", err)
			}

			defer mockPool.Close()

			repo := repository.NewUserRepository(mockPool)

			tt.expect(mockPool)

			got, err := repo.ExistsByEmail(context.Background(), tt.user.Email)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("No error expected in %v but got %v", tt.name, err)
			}

			if err == nil && got != tt.exists {
				t.Errorf("ExistsByEmail() = %v, want %v", got, tt.exists)
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
