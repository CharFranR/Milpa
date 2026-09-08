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
		name      string
		ctx       context.Context
		companyID uuid.UUID
		rating    int
		comment   string
		saveErr   error
		wantErr   error
	}{
		{name: "happy path", ctx: principalCtx(), companyID: testCompanyID, rating: 5, comment: "Great quality"},
		{name: "unauthenticated", ctx: context.Background(), companyID: testCompanyID, rating: 5, wantErr: auth.ErrUnauthenticated},
		{name: "rating too low", ctx: principalCtx(), companyID: testCompanyID, rating: 0, wantErr: domain.ErrInvalidRating},
		{name: "rating too high", ctx: principalCtx(), companyID: testCompanyID, rating: 6, wantErr: domain.ErrInvalidRating},
		{name: "save error", ctx: principalCtx(), companyID: testCompanyID, rating: 5, saveErr: errFake, wantErr: errFake},
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

			got, err := uc.CreateReview(tt.ctx, dto.CreateReviewRequest{CompanyID: tt.companyID, Rating: tt.rating, Comment: tt.comment})

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
			if saved.UserID != testUserID {
				t.Errorf("saved user id = %v, want %v", saved.UserID, testUserID)
			}
			if saved.Rating != tt.rating {
				t.Errorf("saved rating = %d, want %d", saved.Rating, tt.rating)
			}
		})
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
				if dto.UserID != tt.reviews[i].UserID {
					t.Errorf("dto %d user id = %v, want %v", i, dto.UserID, tt.reviews[i].UserID)
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
