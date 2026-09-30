package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestReviewUseCaseCreateReview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		companyID  uuid.UUID
		targetType string
		targetID   uuid.UUID
		rating     int
		comment    string
		saveErr    error
		wantErr    error
	}{
		{name: "happy path", ctx: principalCtx(), companyID: testCompanyID, rating: 5, comment: "Great quality"},
		{name: "unauthenticated", ctx: context.Background(), companyID: testCompanyID, rating: 5, wantErr: auth.ErrUnauthenticated},
		{name: "rating too low", ctx: principalCtx(), companyID: testCompanyID, rating: 0, wantErr: domain.ErrInvalidRating},
		{name: "rating too high", ctx: principalCtx(), companyID: testCompanyID, rating: 6, wantErr: domain.ErrInvalidRating},
		{name: "save error", ctx: principalCtx(), companyID: testCompanyID, rating: 5, saveErr: errFake, wantErr: errFake},
		{name: "no target at all", ctx: principalCtx(), rating: 4, wantErr: domain.ErrTargetRequired},
		{
			name: "company and target together is ambiguous", ctx: principalCtx(), companyID: testCompanyID,
			targetType: "user", targetID: testOtherID, rating: 4, wantErr: domain.ErrReviewTargetMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := newFakeReviewRepo()
			if tt.saveErr != nil {
				reviewRepo.save = func(ctx context.Context, review *domain.Review) error {
					return tt.saveErr
				}
			}
			uc := usecases.NewReviewUseCase(reviewRepo, newFakeTimer())

			got, err := uc.CreateReview(tt.ctx, dto.CreateReviewRequest{
				CompanyID:  tt.companyID,
				TargetType: tt.targetType,
				TargetID:   tt.targetID,
				Rating:     tt.rating,
				Comment:    tt.comment,
			})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == uuid.Nil {
				t.Error("expected a generated ID, got nil UUID")
			}
			if got.UserID != testUserID {
				t.Errorf("user id = %v, want %v", got.UserID, testUserID)
			}
			if got.CompanyID != tt.companyID {
				t.Errorf("company id = %v, want %v", got.CompanyID, tt.companyID)
			}
			if got.Rating != tt.rating {
				t.Errorf("rating = %d, want %d", got.Rating, tt.rating)
			}
			if got.Comment != tt.comment {
				t.Errorf("comment = %q, want %q", got.Comment, tt.comment)
			}
			if !got.CreatedAt.Equal(fixedTime) {
				t.Errorf("created at = %v, want %v", got.CreatedAt, fixedTime)
			}
			if len(reviewRepo.saved) != 1 {
				t.Fatalf("saved reviews = %d, want 1", len(reviewRepo.saved))
			}
			saved := reviewRepo.saved[0]
			if saved.AuthorID != testUserID {
				t.Errorf("saved author id = %v, want %v", saved.AuthorID, testUserID)
			}
			if saved.Rating != tt.rating {
				t.Errorf("saved rating = %d, want %d", saved.Rating, tt.rating)
			}
		})
	}
}

// TestReviewUseCaseCreateReviewTargets covers the RF-15 half: a review is about
// a target of a type, and the two request shapes are the two ways of saying so.
func TestReviewUseCaseCreateReviewTargets(t *testing.T) {
	t.Parallel()

	farmerID := uuid.New()

	tests := []struct {
		name        string
		req         dto.CreateReviewRequest
		wantType    domain.ReviewTargetType
		wantTarget  uuid.UUID
		wantCompany uuid.UUID
		wantErr     error
	}{
		{
			name:        "company_id is the company target",
			req:         dto.CreateReviewRequest{CompanyID: testCompanyID, Rating: 4},
			wantType:    domain.ReviewTargetCompany,
			wantTarget:  testCompanyID,
			wantCompany: testCompanyID,
		},
		{
			name:        "explicit company target mirrors the company",
			req:         dto.CreateReviewRequest{TargetType: "company", TargetID: testCompanyID, Rating: 4},
			wantType:    domain.ReviewTargetCompany,
			wantTarget:  testCompanyID,
			wantCompany: testCompanyID,
		},
		{
			name:       "a user target carries no company",
			req:        dto.CreateReviewRequest{TargetType: "user", TargetID: farmerID, Rating: 4},
			wantType:   domain.ReviewTargetUser,
			wantTarget: farmerID,
		},
		{name: "unknown target type", req: dto.CreateReviewRequest{TargetType: "transaction", TargetID: farmerID, Rating: 4}, wantErr: domain.ErrInvalidReviewTargetType},
		{name: "self review", req: dto.CreateReviewRequest{TargetType: "user", TargetID: testUserID, Rating: 4}, wantErr: domain.ErrSelfReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := newFakeReviewRepo()
			uc := usecases.NewReviewUseCase(reviewRepo, newFakeTimer())

			got, err := uc.CreateReview(principalCtx(), tt.req)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if len(reviewRepo.saved) != 0 {
					t.Fatalf("saved reviews = %d, want 0", len(reviewRepo.saved))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TargetType != string(tt.wantType) || got.TargetID != tt.wantTarget {
				t.Errorf("dto target = %q/%v, want %q/%v", got.TargetType, got.TargetID, tt.wantType, tt.wantTarget)
			}
			if got.CompanyID != tt.wantCompany {
				t.Errorf("dto company = %v, want %v", got.CompanyID, tt.wantCompany)
			}
			saved := reviewRepo.saved[0]
			if saved.TargetType != tt.wantType || saved.TargetID != tt.wantTarget {
				t.Errorf("saved target = %q/%v, want %q/%v", saved.TargetType, saved.TargetID, tt.wantType, tt.wantTarget)
			}
			if saved.CompanyID != tt.wantCompany {
				t.Errorf("saved company = %v, want %v", saved.CompanyID, tt.wantCompany)
			}
		})
	}
}

// TestReviewUseCaseGetAverageRating is the use-case half of the RF-15 aggregate:
// the count travels with the average, because 5.0 from two reviews and 5.0 from
// two hundred are not the same claim.
func TestReviewUseCaseGetAverageRating(t *testing.T) {
	t.Parallel()

	reviewRepo := newFakeReviewRepo()
	uc := usecases.NewReviewUseCase(reviewRepo, newFakeTimer())

	got, err := uc.GetAverageRating(context.Background(), domain.ReviewTargetCompany, testCompanyID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Average != 4.5 || got.Count != 2 {
		t.Errorf("average/count = %v/%d, want 4.5/2", got.Average, got.Count)
	}
	if got.TargetType != string(domain.ReviewTargetCompany) || got.TargetID != testCompanyID {
		t.Errorf("dto target = %q/%v, want company/%v", got.TargetType, got.TargetID, testCompanyID)
	}

	for _, bad := range []struct {
		name       string
		targetType domain.ReviewTargetType
		targetID   uuid.UUID
		wantErr    error
	}{
		{name: "unknown target type", targetType: "transaction", targetID: testCompanyID, wantErr: domain.ErrInvalidReviewTargetType},
		{name: "no target", targetType: domain.ReviewTargetUser, targetID: uuid.Nil, wantErr: domain.ErrTargetRequired},
	} {
		t.Run(bad.name, func(t *testing.T) {
			if _, err := uc.GetAverageRating(context.Background(), bad.targetType, bad.targetID); !errors.Is(err, bad.wantErr) {
				t.Fatalf("error = %v, want %v", err, bad.wantErr)
			}
		})
	}

	reviewRepo.averageRating = func(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) (float64, int, error) {
		return 0, 0, errFake
	}
	if _, err := uc.GetAverageRating(context.Background(), domain.ReviewTargetUser, testOtherID); !errors.Is(err, errFake) {
		t.Fatalf("error = %v, want %v", err, errFake)
	}
}

func TestReviewUseCaseFindByUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reviews []domain.Review
		repoErr error
		wantLen int
		wantErr error
	}{
		{
			name: "happy path",
			reviews: func() []domain.Review {
				first := *mustReview()
				second := *mustReview()
				second.ID = testOtherID
				second.Rating = 4
				second.Comment = "Good enough"
				return []domain.Review{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", reviews: []domain.Review{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := newFakeReviewRepo()
			if tt.repoErr != nil {
				reviewRepo.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Review, error) {
					return nil, tt.repoErr
				}
			} else {
				reviewRepo.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Review, error) {
					return tt.reviews, nil
				}
			}
			uc := usecases.NewReviewUseCase(reviewRepo, newFakeTimer())

			got, err := uc.FindByUser(context.Background(), testUserID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("dtos = %d, want %d", len(got), tt.wantLen)
			}
			for i, dto := range got {
				if dto.ID != tt.reviews[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.reviews[i].ID)
				}
				if dto.UserID != tt.reviews[i].AuthorID {
					t.Errorf("dto %d user id = %v, want %v", i, dto.UserID, tt.reviews[i].AuthorID)
				}
				if dto.CompanyID != tt.reviews[i].CompanyID {
					t.Errorf("dto %d company id = %v, want %v", i, dto.CompanyID, tt.reviews[i].CompanyID)
				}
				if dto.Rating != tt.reviews[i].Rating {
					t.Errorf("dto %d rating = %d, want %d", i, dto.Rating, tt.reviews[i].Rating)
				}
				if dto.Comment != tt.reviews[i].Comment {
					t.Errorf("dto %d comment = %q, want %q", i, dto.Comment, tt.reviews[i].Comment)
				}
			}
		})
	}
}

func TestReviewUseCaseFindByCompany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reviews []domain.Review
		repoErr error
		wantLen int
		wantErr error
	}{
		{
			name: "happy path",
			reviews: func() []domain.Review {
				first := *mustReview()
				second := *mustReview()
				second.ID = testOtherID
				return []domain.Review{first, second}
			}(),
			wantLen: 2,
		},
		{name: "empty", reviews: []domain.Review{}, wantLen: 0},
		{name: "repo error", repoErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := newFakeReviewRepo()
			if tt.repoErr != nil {
				reviewRepo.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
					return nil, tt.repoErr
				}
			} else {
				reviewRepo.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
					return tt.reviews, nil
				}
			}
			uc := usecases.NewReviewUseCase(reviewRepo, newFakeTimer())

			got, err := uc.FindByCompany(context.Background(), testCompanyID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("dtos = %d, want %d", len(got), tt.wantLen)
			}
			for i, dto := range got {
				if dto.ID != tt.reviews[i].ID {
					t.Errorf("dto %d id = %v, want %v", i, dto.ID, tt.reviews[i].ID)
				}
				if dto.Rating != tt.reviews[i].Rating {
					t.Errorf("dto %d rating = %d, want %d", i, dto.Rating, tt.reviews[i].Rating)
				}
			}
		})
	}
}
