package repository_test

import (
	"context"
	"errors"

	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
)

func TestCategoryFindAll(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description"}).
					AddRow(testCategoryID, "Cat", "desc").
					AddRow(uuid.MustParse("55555555-5555-5555-5555-555555555556"), "Cat2", "desc2")
				m.ExpectQuery("FROM categories").WillReturnRows(rows)
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM categories").WillReturnError(errors.New("query failed"))
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
			repo := repository.NewCategoryRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindAll(context.Background())

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestCategoryFindByID(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description"}).
					AddRow(testCategoryID, "Cat", "desc")
				m.ExpectQuery("FROM categories").WithArgs(testCategoryID).WillReturnRows(rows)
			},
		},
		{
			name:    "not found",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM categories").WithArgs(testCategoryID).WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description"}))
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM categories").WithArgs(testCategoryID).WillReturnError(errors.New("select failed"))
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
			repo := repository.NewCategoryRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByID(context.Background(), testCategoryID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
