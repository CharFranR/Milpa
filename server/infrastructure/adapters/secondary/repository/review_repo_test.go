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

func TestReviewFindByCompany(t *testing.T) {
	reviewID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	review := &domain.Review{
		ID:        reviewID,
		UserID:    testUserID,
		CompanyID: companyID,
		Rating:    5,
		Comment:   "excelente",
		CreatedAt: fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "company_id", "rating", "comment", "created_at"}).
					AddRow(review.ID, review.UserID, review.CompanyID, review.Rating, review.Comment, review.CreatedAt)
				m.ExpectQuery("FROM reviews").WithArgs(companyID).WillReturnRows(rows)
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM reviews").WithArgs(companyID).WillReturnError(errors.New("query failed"))
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
			repo := repository.NewReviewRepository(mockPool)

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

func TestReviewFindByUser(t *testing.T) {
	reviewID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := testUserID

	review := &domain.Review{
		ID:        reviewID,
		UserID:    userID,
		CompanyID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Rating:    4,
		Comment:   "bueno",
		CreatedAt: fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "company_id", "rating", "comment", "created_at"}).
					AddRow(review.ID, review.UserID, review.CompanyID, review.Rating, review.Comment, review.CreatedAt)
				m.ExpectQuery("FROM reviews").WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM reviews").WithArgs(userID).WillReturnError(errors.New("query failed"))
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
			repo := repository.NewReviewRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByUser(context.Background(), userID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestReviewSave(t *testing.T) {
	reviewID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	review := &domain.Review{
		ID:        reviewID,
		UserID:    testUserID,
		CompanyID: companyID,
		Rating:    5,
		Comment:   "excelente",
		CreatedAt: fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO reviews").
					WithArgs(review.ID, review.UserID, review.CompanyID, review.Rating, review.Comment, review.CreatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO reviews").
					WithArgs(review.ID, review.UserID, review.CompanyID, review.Rating, review.Comment, review.CreatedAt).
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
			repo := repository.NewReviewRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Save(context.Background(), review)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
