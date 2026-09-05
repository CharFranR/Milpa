package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
)

var testInquiryUserID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testInquiryOfferingID uuid.UUID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
var testInquiryID uuid.UUID = uuid.MustParse("77777777-7777-7777-7777-777777777777")
var testInquiryID2 uuid.UUID = uuid.MustParse("77777777-7777-7777-7777-777777777778")

func setupInquiryTestData(t *testing.T) {
	t.Helper()

	userRepo := repository.NewUserRepository(TestPool)
	companyRepo := repository.NewCompanyRepository(TestPool)
	offeringRepo := repository.NewOfferingRepository(TestPool)

	_, err := userRepo.Save(context.Background(), &domain.User{
		ID:           testInquiryUserID,
		FirstName:    "Inquirer",
		LastName:     "User",
		Role:         domain.RoleMIPYME,
		Email:        "inquirer@example.com",
		PhoneNumber:  "0000-0000",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	})
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	inquiryCompanyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	err = companyRepo.Save(context.Background(), &domain.Company{
		ID:          inquiryCompanyID,
		Name:        "Inquiry Company",
		Owner:       domain.User{ID: testInquiryUserID},
		Address:     domain.Address{},
		Description: "Company for inquiries",
		PhoneNumber: "1234-5678",
		Email:       "inquiry-company@example.com",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	})
	if err != nil {
		t.Fatalf("insert company: %v", err)
	}

	err = offeringRepo.Save(context.Background(), &domain.Offering{
		ID:        testInquiryOfferingID,
		CompanyID: inquiryCompanyID,
		Type:      domain.OfferingProduct,
		Name:      "Test Offering",
		Price:     100.00,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	})
	if err != nil {
		t.Fatalf("insert offering: %v", err)
	}
}

func TestInquirySave(t *testing.T) {
	setupInquiryTestData(t)
	db := repository.NewInquiryRepository(TestPool)

	tests := []struct {
		Name        string
		ExpectedErr error
		Inquiry     *domain.Inquiry
	}{
		{
			Name: "Happy Path",
			Inquiry: &domain.Inquiry{
				ID:         testInquiryID,
				UserID:     testInquiryUserID,
				OfferingID: testInquiryOfferingID,
				Message:    "Is this available?",
				Status:     domain.InquiryPending,
				CreatedAt:  fixedTime,
			},
		},
		{
			Name: "Happy Path with status",
			Inquiry: &domain.Inquiry{
				ID:         testInquiryID2,
				UserID:     testInquiryUserID,
				OfferingID: testInquiryOfferingID,
				Message:    "What is the delivery time?",
				Status:     domain.InquiryRead,
				CreatedAt:  fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Save(context.Background(), tt.Inquiry)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Save() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Save() unexpected error: %v", err)
			}
		})
	}
}

func TestInquiryFindByID(t *testing.T) {
	setupInquiryTestData(t)
	db := repository.NewInquiryRepository(TestPool)

	saved := &domain.Inquiry{
		ID:         testInquiryID,
		UserID:     testInquiryUserID,
		OfferingID: testInquiryOfferingID,
		Message:    "Find me",
		Status:     domain.InquiryPending,
		CreatedAt:  fixedTime,
	}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ID          uuid.UUID
		ExpectedErr error
	}{
		{
			Name:        "Happy Path",
			ID:          testInquiryID,
			ExpectedErr: nil,
		},
		{
			Name:        "Not Found",
			ID:          uuid.MustParse("77777777-7777-7777-7777-777777777799"),
			ExpectedErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			inquiry, err := db.FindByID(context.Background(), tt.ID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByID() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByID() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByID() unexpected error: %v", err)
				return
			}

			if inquiry == nil {
				t.Fatal("FindByID() returned nil inquiry")
			}
			if inquiry.ID != tt.ID {
				t.Errorf("FindByID() ID = %v, want %v", inquiry.ID, tt.ID)
			}
			if inquiry.Message != saved.Message {
				t.Errorf("FindByID() Message = %v, want %v", inquiry.Message, saved.Message)
			}
		})
	}
}

func TestInquiryFindByUser(t *testing.T) {
	setupInquiryTestData(t)
	db := repository.NewInquiryRepository(TestPool)

	inquiry1 := &domain.Inquiry{
		ID:         testInquiryID,
		UserID:     testInquiryUserID,
		OfferingID: testInquiryOfferingID,
		Message:    "First inquiry",
		Status:     domain.InquiryPending,
		CreatedAt:  fixedTime,
	}
	inquiry2 := &domain.Inquiry{
		ID:         testInquiryID2,
		UserID:     testInquiryUserID,
		OfferingID: testInquiryOfferingID,
		Message:    "Second inquiry",
		Status:     domain.InquiryReplied,
		CreatedAt:  fixedTime,
	}

	if err := db.Save(context.Background(), inquiry1); err != nil {
		t.Fatalf("Save inquiry1: %v", err)
	}
	if err := db.Save(context.Background(), inquiry2); err != nil {
		t.Fatalf("Save inquiry2: %v", err)
	}

	tests := []struct {
		Name        string
		UserID      uuid.UUID
		ExpectedLen int
		ExpectedErr error
	}{
		{
			Name:        "User with inquiries",
			UserID:      testInquiryUserID,
			ExpectedLen: 2,
		},
		{
			Name:        "User with no inquiries",
			UserID:      uuid.MustParse("11111111-1111-1111-1111-111111111199"),
			ExpectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			inquiries, err := db.FindByUser(context.Background(), tt.UserID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByUser() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByUser() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByUser() unexpected error: %v", err)
				return
			}

			if len(inquiries) < tt.ExpectedLen {
				t.Fatalf("FindByUser() got %d inquiries, want at least %d", len(inquiries), tt.ExpectedLen)
			}
		})
	}
}

func TestInquiryUpdate(t *testing.T) {
	setupInquiryTestData(t)
	db := repository.NewInquiryRepository(TestPool)

	saved := &domain.Inquiry{
		ID:         testInquiryID,
		UserID:     testInquiryUserID,
		OfferingID: testInquiryOfferingID,
		Message:    "Update me",
		Status:     domain.InquiryPending,
		CreatedAt:  fixedTime,
	}
	if err := db.Save(context.Background(), saved); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	tests := []struct {
		Name        string
		ExpectedErr error
		update_func func() *domain.Inquiry
	}{
		{
			Name: "Happy path - mark as read",
			update_func: func() *domain.Inquiry {
				return &domain.Inquiry{
					ID:     testInquiryID,
					Status: domain.InquiryRead,
				}
			},
		},
		{
			Name: "Happy path - mark as replied",
			update_func: func() *domain.Inquiry {
				return &domain.Inquiry{
					ID:     testInquiryID,
					Status: domain.InquiryReplied,
				}
			},
		},
		{
			Name: "Happy path - close",
			update_func: func() *domain.Inquiry {
				return &domain.Inquiry{
					ID:     testInquiryID,
					Status: domain.InquiryClosed,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Update(context.Background(), tt.update_func())

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
			}
		})
	}
}
