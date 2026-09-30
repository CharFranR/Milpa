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
		ID:         reviewID,
		AuthorID:   testUserID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   companyID,
		CompanyID:  companyID,
		Rating:     5,
		Comment:    "excelente",
		CreatedAt:  fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "author_id", "target_type", "target_id", "company_id", "rating", "comment", "created_at"}).
					AddRow(review.ID, review.AuthorID, review.TargetType, review.TargetID, &review.CompanyID, review.Rating, review.Comment, review.CreatedAt)
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
		ID:         reviewID,
		AuthorID:   userID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		CompanyID:  uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Rating:     4,
		Comment:    "bueno",
		CreatedAt:  fixedTime,
	}

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "author_id", "target_type", "target_id", "company_id", "rating", "comment", "created_at"}).
					AddRow(review.ID, review.AuthorID, review.TargetType, review.TargetID, &review.CompanyID, review.Rating, review.Comment, review.CreatedAt)
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
		ID:         reviewID,
		AuthorID:   testUserID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   companyID,
		CompanyID:  companyID,
		Rating:     5,
		Comment:    "excelente",
		CreatedAt:  fixedTime,
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
					WithArgs(review.ID, review.AuthorID, review.TargetType, review.TargetID, &review.CompanyID, review.Rating, review.Comment, review.CreatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO reviews").
					WithArgs(review.ID, review.AuthorID, review.TargetType, review.TargetID, &review.CompanyID, review.Rating, review.Comment, review.CreatedAt).
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

// TestReviewAverageRating pins the aggregate's shape, COALESCE included: a
// target with no reviews is a 0 over 0, not an error and not a NULL average.
func TestReviewAverageRating(t *testing.T) {
	companyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name      string
		average   float64
		count     int
		repoErr   error
		wantAvg   float64
		wantCount int
		wantErr   bool
	}{
		{name: "rated target", average: 4.25, count: 4, wantAvg: 4.25, wantCount: 4},
		{name: "unrated target", average: 0, count: 0, wantAvg: 0, wantCount: 0},
		{name: "query fail", repoErr: errors.New("query failed"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("mockPool init failed: %v", err)
			}
			defer mockPool.Close()

			if tt.repoErr != nil {
				mockPool.ExpectQuery("AVG\\(rating\\)").
					WithArgs(domain.ReviewTargetCompany, companyID).
					WillReturnError(tt.repoErr)
			} else {
				rows := pgxmock.NewRows([]string{"avg", "count"}).AddRow(tt.average, tt.count)
				mockPool.ExpectQuery("AVG\\(rating\\)").
					WithArgs(domain.ReviewTargetCompany, companyID).
					WillReturnRows(rows)
			}

			repo := repository.NewReviewRepository(mockPool)
			gotAvg, gotCount, err := repo.AverageRating(context.Background(), domain.ReviewTargetCompany, companyID)

			if tt.wantErr != (err != nil) {
				t.Errorf("AverageRating() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && (gotAvg != tt.wantAvg || gotCount != tt.wantCount) {
				t.Errorf("AverageRating() = %v/%d, want %v/%d", gotAvg, gotCount, tt.wantAvg, tt.wantCount)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
