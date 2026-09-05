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

func TestInquiryFindByID(t *testing.T) {
	inquiryID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	userID := testUserID
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	inquiry := &domain.Inquiry{
		ID:         inquiryID,
		UserID:     userID,
		OfferingID: offeringID,
		Message:    "me interesa",
		Status:     domain.InquiryPending,
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
				rows := pgxmock.NewRows([]string{"id", "user_id", "offering_id", "message", "status", "created_at"}).
					AddRow(inquiry.ID, inquiry.UserID, inquiry.OfferingID, inquiry.Message, inquiry.Status, inquiry.CreatedAt)
				m.ExpectQuery("FROM inquiries").WithArgs(inquiryID).WillReturnRows(rows)
			},
		},
		{
			name:    "select fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM inquiries").WithArgs(inquiryID).WillReturnError(errors.New("select failed"))
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
			repo := repository.NewInquiryRepository(mockPool)

			tt.expect(mockPool)
			_, err = repo.FindByID(context.Background(), inquiryID)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestInquiryFindByUser(t *testing.T) {
	inquiryID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	userID := testUserID
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	inquiry := &domain.Inquiry{
		ID:         inquiryID,
		UserID:     userID,
		OfferingID: offeringID,
		Message:    "me interesa",
		Status:     domain.InquiryPending,
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
				rows := pgxmock.NewRows([]string{"id", "user_id", "offering_id", "message", "status", "created_at"}).
					AddRow(inquiry.ID, inquiry.UserID, inquiry.OfferingID, inquiry.Message, inquiry.Status, inquiry.CreatedAt)
				m.ExpectQuery("FROM inquiries").WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:    "query fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("FROM inquiries").WithArgs(userID).WillReturnError(errors.New("query failed"))
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
			repo := repository.NewInquiryRepository(mockPool)

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

func TestInquirySave(t *testing.T) {
	inquiryID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	userID := testUserID
	offeringID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	inquiry := &domain.Inquiry{
		ID:         inquiryID,
		UserID:     userID,
		OfferingID: offeringID,
		Message:    "me interesa",
		Status:     domain.InquiryPending,
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
				m.ExpectExec("INSERT INTO inquiries").
					WithArgs(inquiry.ID, inquiry.UserID, inquiry.OfferingID, inquiry.Message, inquiry.Status, inquiry.CreatedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("INSERT INTO inquiries").
					WithArgs(inquiry.ID, inquiry.UserID, inquiry.OfferingID, inquiry.Message, inquiry.Status, inquiry.CreatedAt).
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
			repo := repository.NewInquiryRepository(mockPool)

			tt.expect(mockPool)
			err = repo.Save(context.Background(), inquiry)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}

func TestInquiryUpdate(t *testing.T) {
	inquiryID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	tests := []struct {
		name    string
		wantErr bool
		expect  func(m pgxmock.PgxPoolIface)
	}{
		{
			name: "Happy path",
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("UPDATE inquiries").WithArgs(domain.InquiryRead, inquiryID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name:    "exec fail",
			wantErr: true,
			expect: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec("UPDATE inquiries").WithArgs(domain.InquiryRead, inquiryID).WillReturnError(errors.New("exec failed"))
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
			repo := repository.NewInquiryRepository(mockPool)

			inquiry := &domain.Inquiry{ID: inquiryID, Status: domain.InquiryRead}
			tt.expect(mockPool)
			err = repo.Update(context.Background(), inquiry)

			if (tt.wantErr) != (err != nil) {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("in %v expectations were unfulfilled: %v", tt.name, err)
			}
		})
	}
}
