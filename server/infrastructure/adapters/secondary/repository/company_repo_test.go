package repository_test

import (
	"context"
	"errors"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
)

var testCompanyID uuid.UUID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
var testCategoryID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555555")

var companyColumns = []string{
	"id", "name", "owner_id", "description", "phone_number", "email", "website",
	"verified", "created_at", "updated_at",
	"a.id", "a.department", "a.municipality", "a.address_line", "a.latitude", "a.longitude",
}

var categoryColumns = []string{"id", "name", "description"}

func companyRow(id, owner, addressID uuid.UUID) *pgxmock.Rows {
	return pgxmock.NewRows(companyColumns).AddRow(
		id, "Mi Empresa", owner, "una descripcion", "1234-5678", "e@mail.com", "http://web",
		true, fixedTime, fixedTime,
		addressID, "Leon", "Leon", "Calle Central", 12.435, -86.878,
	)
}

func categoryRow(id uuid.UUID) *pgxmock.Rows {
	return pgxmock.NewRows(categoryColumns).AddRow(id, "Cat", "desc")
}

func TestCompanyFindByID(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM companies").WithArgs(testCompanyID).WillReturnRows(companyRow(testCompanyID, testUserID, testAddressID))
				m.ExpectQuery("FROM categories").WithArgs(testCompanyID).WillReturnRows(categoryRow(testCategoryID))
			},
		},
		{
			name:    "not found",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM companies").WithArgs(testCompanyID).WillReturnRows(pgxmock.NewRows(companyColumns))
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM companies").WithArgs(testCompanyID).WillReturnError(errors.New("select failed"))
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
			repo := repository.NewCompanyRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByID(context.Background(), testCompanyID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestCompanyFindByOwner(t *testing.T) {
	companyID2 := uuid.MustParse("22222222-2222-2222-2222-222222222223")

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows(companyColumns).
					AddRow(testCompanyID, "Mi Empresa", testUserID, "una descripcion", "1234-5678", "e@mail.com", "http://web",
						true, fixedTime, fixedTime, testAddressID, "Leon", "Leon", "Calle Central", 12.435, -86.878).
					AddRow(companyID2, "Otra", testUserID, "desc2", "9999", "o@mail.com", "http://o",
						false, fixedTime, fixedTime, testAddressID, "Leon", "Leon", "Calle 2", 12.4, -86.8)
				m.ExpectQuery("FROM companies").WithArgs(testUserID).WillReturnRows(rows)
				m.ExpectQuery("FROM categories").WithArgs(testCompanyID).WillReturnRows(categoryRow(testCategoryID))
				m.ExpectQuery("FROM categories").WithArgs(companyID2).WillReturnRows(categoryRow(testCategoryID))
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM companies").WithArgs(testUserID).WillReturnError(errors.New("query failed"))
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
			repo := repository.NewCompanyRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByOwner(context.Background(), testUserID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByOwner() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestCompanySave(t *testing.T) {
	cat := domain.Category{ID: testCategoryID, Name: "Cat", Description: "desc"}

	withAddress := &domain.Company{
		ID:          testCompanyID,
		Name:        "Mi Empresa",
		Owner:       domain.User{ID: testUserID},
		Category:    []domain.Category{cat},
		Address:     domain.Address{Department: "Leon", Municipality: "Leon", AddressLine: "Calle Central", Latitude: 12.435, Longitude: -86.878},
		Description: "una descripcion",
		PhoneNumber: "1234-5678",
		Email:       "e@mail.com",
		Website:     "http://web",
		Verified:    true,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	withoutAddress := &domain.Company{
		ID:          testCompanyID,
		Name:        "Mi Empresa",
		Owner:       domain.User{ID: testUserID},
		Category:    []domain.Category{cat},
		Address:     domain.Address{},
		Description: "una descripcion",
		PhoneNumber: "1234-5678",
		Email:       "e@mail.com",
		Website:     "http://web",
		Verified:    true,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		company *domain.Company
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name:    "Happy path with address",
			wantErr: false,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(testAddressID.String()))
				m.ExpectExec("INSERT INTO companies").
					WithArgs(withAddress.ID, withAddress.Name, withAddress.Owner.ID, &testAddressID, withAddress.Description, withAddress.PhoneNumber, withAddress.Email, withAddress.Website, withAddress.Verified, withAddress.CreatedAt, withAddress.UpdatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "Happy path without address",
			wantErr: false,
			company: withoutAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO companies").
					WithArgs(withoutAddress.ID, withoutAddress.Name, withoutAddress.Owner.ID, nullID, withoutAddress.Description, withoutAddress.PhoneNumber, withoutAddress.Email, withoutAddress.Website, withoutAddress.Verified, withoutAddress.CreatedAt, withoutAddress.UpdatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withoutAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "begin fail",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("db begin failed"))
			},
		},
		{
			name:    "address insert fails",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude).
					WillReturnError(errors.New("address failed"))
				m.ExpectRollback()
			},
		},
		{
			name:    "company insert fails",
			wantErr: true,
			company: withoutAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO companies").
					WithArgs(withoutAddress.ID, withoutAddress.Name, withoutAddress.Owner.ID, nullID, withoutAddress.Description, withoutAddress.PhoneNumber, withoutAddress.Email, withoutAddress.Website, withoutAddress.Verified, withoutAddress.CreatedAt, withoutAddress.UpdatedAt).
					WillReturnError(errors.New("company failed"))
				m.ExpectRollback()
			},
		},
		{
			name:    "commit fail",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery("INSERT INTO addresses").
					WithArgs(pgxmock.AnyArg(), withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(testAddressID.String()))
				m.ExpectExec("INSERT INTO companies").
					WithArgs(withAddress.ID, withAddress.Name, withAddress.Owner.ID, &testAddressID, withAddress.Description, withAddress.PhoneNumber, withAddress.Email, withAddress.Website, withAddress.Verified, withAddress.CreatedAt, withAddress.UpdatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("db commit failed"))
				m.ExpectRollback()
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
			repo := repository.NewCompanyRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Save(context.Background(), tt.company)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestCompanyUpdate(t *testing.T) {
	cat := domain.Category{ID: testCategoryID, Name: "Cat", Description: "desc"}

	withAddress := &domain.Company{
		ID:          testCompanyID,
		Name:        "Mi Empresa",
		Owner:       domain.User{ID: testUserID},
		Category:    []domain.Category{cat},
		Address:     domain.Address{ID: testAddressID, Department: "Leon", Municipality: "Leon", AddressLine: "Calle Central", Latitude: 12.435, Longitude: -86.878},
		Description: "una descripcion",
		PhoneNumber: "1234-5678",
		Email:       "e@mail.com",
		Website:     "http://web",
		Verified:    true,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime2,
	}

	withoutAddress := &domain.Company{
		ID:          testCompanyID,
		Name:        "Mi Empresa",
		Owner:       domain.User{ID: testUserID},
		Category:    []domain.Category{cat},
		Address:     domain.Address{},
		Description: "una descripcion",
		PhoneNumber: "1234-5678",
		Email:       "e@mail.com",
		Website:     "http://web",
		Verified:    true,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime2,
	}

	tests := []struct {
		name    string
		wantErr bool
		company *domain.Company
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name:    "Happy path with address",
			wantErr: false,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE addresses").
					WithArgs(withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude, withAddress.Address.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectExec("UPDATE companies").
					WithArgs(withAddress.Name, withAddress.Description, withAddress.PhoneNumber, withAddress.Email, withAddress.Website, withAddress.Verified, withAddress.UpdatedAt, &testAddressID, withAddress.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectExec("DELETE FROM company_categories").WithArgs(withAddress.ID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "Happy path without address",
			wantErr: false,
			company: withoutAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE companies").
					WithArgs(withoutAddress.Name, withoutAddress.Description, withoutAddress.PhoneNumber, withoutAddress.Email, withoutAddress.Website, withoutAddress.Verified, withoutAddress.UpdatedAt, nullID, withoutAddress.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectExec("DELETE FROM company_categories").WithArgs(withoutAddress.ID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withoutAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
		},
		{
			name:    "begin fail",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("db begin failed"))
			},
		},
		{
			name:    "address update fails",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE addresses").
					WithArgs(withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude, withAddress.Address.ID).
					WillReturnError(errors.New("address update failed"))
				m.ExpectRollback()
			},
		},
		{
			name:    "company update fails",
			wantErr: true,
			company: withoutAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE companies").
					WithArgs(withoutAddress.Name, withoutAddress.Description, withoutAddress.PhoneNumber, withoutAddress.Email, withoutAddress.Website, withoutAddress.Verified, withoutAddress.UpdatedAt, nullID, withoutAddress.ID).
					WillReturnError(errors.New("company update failed"))
				m.ExpectRollback()
			},
		},
		{
			name:    "commit fail",
			wantErr: true,
			company: withAddress,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE addresses").
					WithArgs(withAddress.Address.Department, withAddress.Address.Municipality, withAddress.Address.AddressLine, withAddress.Address.Latitude, withAddress.Address.Longitude, withAddress.Address.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectExec("UPDATE companies").
					WithArgs(withAddress.Name, withAddress.Description, withAddress.PhoneNumber, withAddress.Email, withAddress.Website, withAddress.Verified, withAddress.UpdatedAt, &testAddressID, withAddress.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectExec("DELETE FROM company_categories").WithArgs(withAddress.ID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO company_categories").WithArgs(withAddress.ID, cat.ID).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("db commit failed"))
				m.ExpectRollback()
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
			repo := repository.NewCompanyRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Update(context.Background(), tt.company)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
