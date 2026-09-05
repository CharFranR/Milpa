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

func TestOfferingFindByID(t *testing.T) {
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	offering := &domain.Offering{
		ID:          offeringID,
		CompanyID:   companyID,
		Type:        domain.OfferingProduct,
		Name:        "maiz",
		Description: "maiz organico",
		Price:       19.99,
		ImageURL:    "http://img",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "company_id", "type", "name", "description", "price", "image_url", "created_at", "updated_at"}).
					AddRow(offering.ID, offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.CreatedAt, offering.UpdatedAt)
				m.ExpectQuery("FROM offerings").WithArgs(offeringID).WillReturnRows(rows)
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM offerings").WithArgs(offeringID).WillReturnError(errors.New("select failed"))
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
			repo := repository.NewOfferingRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByID(context.Background(), offeringID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestOfferingFindByCompany(t *testing.T) {
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	offering := &domain.Offering{
		ID:          offeringID,
		CompanyID:   companyID,
		Type:        domain.OfferingProduct,
		Name:        "maiz",
		Description: "maiz organico",
		Price:       19.99,
		ImageURL:    "http://img",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "company_id", "type", "name", "description", "price", "image_url", "created_at", "updated_at"}).
					AddRow(offering.ID, offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.CreatedAt, offering.UpdatedAt)
				m.ExpectQuery("FROM offerings").WithArgs(companyID).WillReturnRows(rows)
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM offerings").WithArgs(companyID).WillReturnError(errors.New("query failed"))
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
			repo := repository.NewOfferingRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByCompany(context.Background(), companyID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByCompany() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestOfferingSave(t *testing.T) {
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	offering := &domain.Offering{
		ID:          offeringID,
		CompanyID:   companyID,
		Type:        domain.OfferingProduct,
		Name:        "maiz",
		Description: "maiz organico",
		Price:       19.99,
		ImageURL:    "http://img",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO offerings").
					WithArgs(offering.ID, offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.CreatedAt, offering.UpdatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO offerings").
					WithArgs(offering.ID, offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.CreatedAt, offering.UpdatedAt).
					WillReturnError(errors.New("exec failed"))
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
			repo := repository.NewOfferingRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Save(context.Background(), offering)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestOfferingUpdate(t *testing.T) {
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	offering := &domain.Offering{
		ID:          offeringID,
		CompanyID:   companyID,
		Type:        domain.OfferingProduct,
		Name:        "maiz",
		Description: "maiz organico",
		Price:       19.99,
		ImageURL:    "http://img",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("UPDATE offerings").
					WithArgs(offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.UpdatedAt, offering.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("UPDATE offerings").
					WithArgs(offering.CompanyID, offering.Type, offering.Name, offering.Description, offering.Price, offering.ImageURL, offering.UpdatedAt, offering.ID).
					WillReturnError(errors.New("exec failed"))
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
			repo := repository.NewOfferingRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Update(context.Background(), offering)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestOfferingDelete(t *testing.T) {
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("DELETE FROM offerings").WithArgs(offeringID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("DELETE FROM offerings").WithArgs(offeringID).WillReturnError(errors.New("exec failed"))
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
			repo := repository.NewOfferingRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Delete(context.Background(), offeringID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
