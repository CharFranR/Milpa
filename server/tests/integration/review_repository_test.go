package integration

import (
	"context"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"testing"

	"github.com/google/uuid"
)

var testReviewUserID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testReviewCompanyID uuid.UUID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
var testReviewID uuid.UUID = uuid.MustParse("66666666-6666-6666-6666-666666666666")
var testReviewID2 uuid.UUID = uuid.MustParse("66666666-6666-6666-6666-666666666667")

func setupReviewTestData(t *testing.T) {
	t.Helper()

	userRepo := repository.NewUserRepository(TestPool)
	companyRepo := repository.NewCompanyRepository(TestPool)

	_, err := userRepo.Save(context.Background(), &domain.User{
		ID:           testReviewUserID,
		FirstName:    "Reviewer",
		LastName:     "User",
		Role:         domain.RoleMIPYME,
		Email:        "reviewer@example.com",
		PhoneNumber:  "0000-0000",
		PasswordHash: "hash",
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	})
	if err != nil {
		t.Fatalf("insert reviewer user: %v", err)
	}

	err = companyRepo.Save(context.Background(), &domain.Company{
		ID:          testReviewCompanyID,
		Name:        "Review Company",
		Owner:       domain.User{ID: testReviewUserID},
		Address:     domain.Address{},
		Description: "Company for reviews",
		PhoneNumber: "1234-5678",
		Email:       "review-company@example.com",
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	})
	if err != nil {
		t.Fatalf("insert company: %v", err)
	}
}

func TestReviewSave(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)

	tests := []struct {
		Name        string
		ExpectedErr error
		Review      *domain.Review
	}{
		{
			Name: "Happy Path",
			Review: &domain.Review{
				ID:        testReviewID,
				UserID:    testReviewUserID,
				CompanyID: testReviewCompanyID,
				Rating:    5,
				Comment:   "Excellent service!",
				CreatedAt: fixedTime,
			},
		},
		{
			Name: "Happy Path rating 1",
			Review: &domain.Review{
				ID:        testReviewID2,
				UserID:    testReviewUserID,
				CompanyID: testReviewCompanyID,
				Rating:    1,
				Comment:   "Poor experience",
				CreatedAt: fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := db.Save(context.Background(), tt.Review)

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

func TestReviewFindByCompany(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)

	review1 := &domain.Review{
		ID:        testReviewID,
		UserID:    testReviewUserID,
		CompanyID: testReviewCompanyID,
		Rating:    5,
		Comment:   "Great!",
		CreatedAt: fixedTime,
	}
	review2 := &domain.Review{
		ID:        testReviewID2,
		UserID:    testReviewUserID,
		CompanyID: testReviewCompanyID,
		Rating:    3,
		Comment:   "Okay",
		CreatedAt: fixedTime,
	}

	if err := db.Save(context.Background(), review1); err != nil {
		t.Fatalf("Save review1: %v", err)
	}
	if err := db.Save(context.Background(), review2); err != nil {
		t.Fatalf("Save review2: %v", err)
	}

	tests := []struct {
		Name            string
		CompanyID       uuid.UUID
		ExpectedLen     int
		ExpectedErr     error
		ExpectedComments []string
	}{
		{
			Name:            "Company with reviews",
			CompanyID:       testReviewCompanyID,
			ExpectedLen:     2,
			ExpectedComments: []string{"Great!", "Okay"},
		},
		{
			Name:        "Company with no reviews",
			CompanyID:   uuid.MustParse("33333333-3333-3333-3333-333333333399"),
			ExpectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reviews, err := db.FindByCompany(context.Background(), tt.CompanyID)

			if tt.ExpectedErr != nil {
				if err == nil {
					t.Errorf("FindByCompany() error = nil, wantErr %v", tt.ExpectedErr)
				} else if !errors.Is(err, tt.ExpectedErr) {
					t.Errorf("FindByCompany() error = %v, wantErr %v", err, tt.ExpectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("FindByCompany() unexpected error: %v", err)
				return
			}

			if len(reviews) < tt.ExpectedLen {
				t.Fatalf("FindByCompany() got %d reviews, want at least %d", len(reviews), tt.ExpectedLen)
			}

			if tt.ExpectedComments != nil {
				found := make(map[string]bool)
				for _, r := range reviews {
					found[r.Comment] = true
				}
				for _, comment := range tt.ExpectedComments {
					if !found[comment] {
						t.Errorf("FindByCompany() did not find review '%s'", comment)
					}
				}
			}
		})
	}
}

func TestReviewFindByUser(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)

	review := &domain.Review{
		ID:        testReviewID,
		UserID:    testReviewUserID,
		CompanyID: testReviewCompanyID,
		Rating:    4,
		Comment:   "User review",
		CreatedAt: fixedTime,
	}

	if err := db.Save(context.Background(), review); err != nil {
		t.Fatalf("Save: %v", err)
	}

	tests := []struct {
		Name        string
		UserID      uuid.UUID
		ExpectedLen int
		ExpectedErr error
	}{
		{
			Name:        "User with reviews",
			UserID:      testReviewUserID,
			ExpectedLen: 1,
		},
		{
			Name:        "User with no reviews",
			UserID:      uuid.MustParse("11111111-1111-1111-1111-111111111199"),
			ExpectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reviews, err := db.FindByUser(context.Background(), tt.UserID)

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

			if len(reviews) < tt.ExpectedLen {
				t.Fatalf("FindByUser() got %d reviews, want at least %d", len(reviews), tt.ExpectedLen)
			}

			if tt.ExpectedLen > 0 && reviews[0].UserID != tt.UserID {
				t.Errorf("FindByUser() UserID = %v, want %v", reviews[0].UserID, tt.UserID)
			}
		})
	}
}
