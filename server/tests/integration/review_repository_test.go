package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
)

var testReviewUserID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testReviewCompanyID uuid.UUID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
var testReviewID uuid.UUID = uuid.MustParse("66666666-6666-6666-6666-666666666666")
var testReviewID2 uuid.UUID = uuid.MustParse("66666666-6666-6666-6666-666666666667")
var testReviewUserID2 uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111112")
var testReviewUserID3 uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111113")
var testReviewFarmerID uuid.UUID = uuid.MustParse("11111111-1111-1111-1111-111111111114")

func setupReviewTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)
	companyRepo := repository.NewCompanyRepository(TestPool)

	// Four users, because uq_reviews_author_target means two reviews of the
	// same company have to come from two different authors.
	for _, u := range []*domain.User{
		{ID: testReviewUserID, FirstName: "Reviewer", LastName: "User", Email: "reviewer@example.com"},
		{ID: testReviewUserID2, FirstName: "Second", LastName: "Reviewer", Email: "reviewer-2@example.com"},
		{ID: testReviewUserID3, FirstName: "Third", LastName: "Reviewer", Email: "reviewer-3@example.com"},
		{ID: testReviewFarmerID, FirstName: "Reviewed", LastName: "Farmer", Email: "farmer@example.com"},
	} {
		u.Role = domain.RoleCompradorMinorista
		u.PhoneNumber = "0000-0000"
		u.PasswordHash = "hash"
		u.CreatedAt = fixedTime
		u.UpdatedAt = fixedTime
		if _, err := userRepo.Save(context.Background(), u); err != nil {
			t.Fatalf("insert user %s: %v", u.Email, err)
		}
	}

	err := companyRepo.Save(context.Background(), &domain.Company{
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
				ID:         testReviewID,
				TransactionID: uuid.New(),
				AuthorID:   testReviewUserID,
				TargetType: domain.ReviewTargetCompany,
				TargetID:   testReviewCompanyID,
				CompanyID:  testReviewCompanyID,
				Rating:     5,
				Comment:    "Excellent service!",
				CreatedAt:  fixedTime,
			},
		},
		{
			Name: "Happy Path rating 1",
			Review: &domain.Review{
				ID:         testReviewID2,
				TransactionID: uuid.New(),
				AuthorID:   testReviewUserID2,
				TargetType: domain.ReviewTargetCompany,
				TargetID:   testReviewCompanyID,
				CompanyID:  testReviewCompanyID,
				Rating:     1,
				Comment:    "Poor experience",
				CreatedAt:  fixedTime,
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
		ID:         testReviewID,
		TransactionID: uuid.New(),
		AuthorID:   testReviewUserID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   testReviewCompanyID,
		CompanyID:  testReviewCompanyID,
		Rating:     5,
		Comment:    "Great!",
		CreatedAt:  fixedTime,
	}
	review2 := &domain.Review{
		ID:         testReviewID2,
		TransactionID: uuid.New(),
		AuthorID:   testReviewUserID2,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   testReviewCompanyID,
		CompanyID:  testReviewCompanyID,
		Rating:     3,
		Comment:    "Okay",
		CreatedAt:  fixedTime,
	}

	if err := db.Save(context.Background(), review1); err != nil {
		t.Fatalf("Save review1: %v", err)
	}
	if err := db.Save(context.Background(), review2); err != nil {
		t.Fatalf("Save review2: %v", err)
	}

	tests := []struct {
		Name             string
		CompanyID        uuid.UUID
		ExpectedLen      int
		ExpectedErr      error
		ExpectedComments []string
	}{
		{
			Name:             "Company with reviews",
			CompanyID:        testReviewCompanyID,
			ExpectedLen:      2,
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
		ID:         testReviewID,
		TransactionID: uuid.New(),
		AuthorID:   testReviewUserID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   testReviewCompanyID,
		CompanyID:  testReviewCompanyID,
		Rating:     4,
		Comment:    "User review",
		CreatedAt:  fixedTime,
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

			if tt.ExpectedLen > 0 && reviews[0].AuthorID != tt.UserID {
				t.Errorf("FindByUser() AuthorID = %v, want %v", reviews[0].AuthorID, tt.UserID)
			}
		})
	}
}

// TestReviewSaveUserTarget is the RF-15 capability the table could not express
// before: a review of a farmer, with no company to mirror.
func TestReviewSaveUserTarget(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)
	ctx := context.Background()

	saved, err := domain.NewReview(testReviewUserID, domain.ReviewTargetUser, testReviewFarmerID, uuid.Nil, 4, "Reliable delivery", fixedTime, uuid.New())
	if err != nil {
		t.Fatalf("NewReview() error: %v", err)
	}
	if err := db.Save(ctx, saved); err != nil {
		t.Fatalf("Save() a user review: %v", err)
	}

	// Read the row back with raw SQL: this tests what the migration actually
	// stored, and a repository read that projected away the columns under test
	// would prove nothing.
	var authorID, targetID uuid.UUID
	var companyID *uuid.UUID
	var targetType string
	if err := TestPool.QueryRow(ctx,
		`SELECT author_id, target_type, target_id, company_id FROM reviews WHERE id = $1`, saved.ID,
	).Scan(&authorID, &targetType, &targetID, &companyID); err != nil {
		t.Fatalf("read back the user review: %v", err)
	}
	if authorID != testReviewUserID || targetID != testReviewFarmerID {
		t.Errorf("author/target = %v/%v, want %v/%v", authorID, targetID, testReviewUserID, testReviewFarmerID)
	}
	if targetType != string(domain.ReviewTargetUser) {
		t.Errorf("target_type = %q, want %q", targetType, domain.ReviewTargetUser)
	}
	if companyID != nil {
		t.Errorf("company_id = %v, want NULL for a user review", *companyID)
	}

	// The company listing must not pick it up: it has no company, and
	// GET /reviews?company_id= must keep answering exactly what it answered
	// before RF-15.
	byCompany, err := db.FindByCompany(ctx, testReviewCompanyID)
	if err != nil {
		t.Fatalf("FindByCompany() error: %v", err)
	}
	if len(byCompany) != 0 {
		t.Errorf("FindByCompany() = %d reviews, want 0 for a user-target review", len(byCompany))
	}
}

// TestReviewCompanyMirrorIsEnforcedInStorage proves the CHECK is load-bearing
// and not a comment: a company review whose mirror names a different company is
// refused by PostgreSQL, not only by NewReview.
func TestReviewCompanyMirrorIsEnforcedInStorage(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)

	err := db.Save(context.Background(), &domain.Review{
		ID:         testReviewID,
		TransactionID: uuid.New(),
		AuthorID:   testReviewUserID,
		TargetType: domain.ReviewTargetCompany,
		TargetID:   testReviewCompanyID,
		CompanyID:  testReviewFarmerID, // the farmer is a user, not the company
		Rating:     5,
		Comment:    "mirror does not match the target",
		CreatedAt:  fixedTime,
	})
	if err == nil {
		t.Fatal("Save() with a mismatched company mirror succeeded, want ck_reviews_company_mirror to reject it")
	}
}

// TestReviewOnePerAuthorPerTarget proves uq_reviews_author_target, without which
// a buyer could post the same rating twice and move an average nobody re-decided.
func TestReviewOnePerAuthorPerTarget(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)
	ctx := context.Background()

	first := &domain.Review{
		ID: testReviewID, TransactionID: uuid.New(), AuthorID: testReviewUserID, TargetType: domain.ReviewTargetCompany,
		TargetID: testReviewCompanyID, CompanyID: testReviewCompanyID, Rating: 1, Comment: "first", CreatedAt: fixedTime,
	}
	if err := db.Save(ctx, first); err != nil {
		t.Fatalf("Save() the first review: %v", err)
	}

	second := *first
	second.ID = testReviewID2
	second.Rating = 5
	second.Comment = "second"
	second.CreatedAt = fixedTime.Add(time.Minute)
	if err := db.Save(ctx, &second); err == nil {
		t.Fatal("Save() a second review of the same target by the same author succeeded, want uq_reviews_author_target to reject it")
	}

	// The same author reviewing a different target stays allowed, which is the
	// point of keying on the triple and not on the company.
	other := first
	other.ID = uuid.New()
	other.TransactionID = uuid.New()
	other.TargetID = testReviewFarmerID
	other.TargetType = domain.ReviewTargetUser
	other.CompanyID = uuid.Nil
	if err := db.Save(ctx, other); err != nil {
		t.Errorf("Save() the same author reviewing a farmer: %v, want it to be allowed", err)
	}
}

// TestReviewAverageRating is the aggregate behind GET /reviews/average.
func TestReviewAverageRating(t *testing.T) {
	setupReviewTestData(t)
	db := repository.NewReviewRepository(TestPool)
	ctx := context.Background()

	for _, r := range []*domain.Review{
		{ID: testReviewID, TransactionID: uuid.New(), AuthorID: testReviewUserID, TargetType: domain.ReviewTargetCompany, TargetID: testReviewCompanyID, CompanyID: testReviewCompanyID, Rating: 5, Comment: "a", CreatedAt: fixedTime},
		{ID: testReviewID2, TransactionID: uuid.New(), AuthorID: testReviewUserID2, TargetType: domain.ReviewTargetCompany, TargetID: testReviewCompanyID, CompanyID: testReviewCompanyID, Rating: 4, Comment: "b", CreatedAt: fixedTime.Add(time.Minute)},
	} {
		if err := db.Save(ctx, r); err != nil {
			t.Fatalf("Save() review: %v", err)
		}
	}

	average, count, err := db.AverageRating(ctx, domain.ReviewTargetCompany, testReviewCompanyID)
	if err != nil {
		t.Fatalf("AverageRating() error: %v", err)
	}
	if average != 4.5 || count != 2 {
		t.Errorf("average/count = %v/%d, want 4.5/2", average, count)
	}

	// A target nobody reviewed is a 0 over 0, not an error and not a NULL.
	average, count, err = db.AverageRating(ctx, domain.ReviewTargetUser, testReviewFarmerID)
	if err != nil {
		t.Fatalf("AverageRating() on an unrated target error: %v", err)
	}
	if average != 0 || count != 0 {
		t.Errorf("average/count = %v/%d, want 0/0", average, count)
	}
}
