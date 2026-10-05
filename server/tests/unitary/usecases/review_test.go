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

func newTestReviewUC(reviewRepo *fakeReviewRepo) *usecases.ReviewUseCaseImpl {
	return newTestReviewUCWithCompany(reviewRepo, newFakeCompanyRepo())
}

func newTestReviewUCWithCompany(reviewRepo *fakeReviewRepo, companyRepo *fakeCompanyRepo) *usecases.ReviewUseCaseImpl {
	txRepo := newFakeTxTransactionRepo()
	txRepo.stored.Status = domain.TransactionCompleted
	matchRepo := newFakeTxMatchRepo()
	offerRepo := newFakeTxOfferRepo()
	offerRepo.stored.SupplierID = testOtherID
	requestRepo := newFakeTxRequestRepo()
	requestRepo.stored.BuyerID = testUserID
	return usecases.NewReviewUseCase(reviewRepo, txRepo, matchRepo, offerRepo, requestRepo, companyRepo, newFakeTimer())
}

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
		{name: "happy path", ctx: principalCtx(), companyID: testOtherCompanyID, rating: 5, comment: "Great quality"},
		{name: "unauthenticated", ctx: context.Background(), companyID: testCompanyID, rating: 5, wantErr: auth.ErrUnauthenticated},
		{name: "rating too low", ctx: principalCtx(), companyID: testOtherCompanyID, rating: 0, wantErr: domain.ErrInvalidRating},
		{name: "rating too high", ctx: principalCtx(), companyID: testOtherCompanyID, rating: 6, wantErr: domain.ErrInvalidRating},
		{name: "save error", ctx: principalCtx(), companyID: testOtherCompanyID, rating: 5, saveErr: errFake, wantErr: errFake},
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
			uc := newTestReviewUC(reviewRepo)

			got, err := uc.CreateReview(tt.ctx, dto.CreateReviewRequest{
				CompanyID:     tt.companyID,
				TargetType:    tt.targetType,
				TargetID:      tt.targetID,
				Rating:        tt.rating,
				Comment:       tt.comment,
				TransactionID: txTestTransactID,
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
			req:         dto.CreateReviewRequest{CompanyID: testOtherCompanyID, Rating: 4, TransactionID: txTestTransactID},
			wantType:    domain.ReviewTargetCompany,
			wantTarget:  testOtherCompanyID,
			wantCompany: testOtherCompanyID,
		},
		{
			name:        "explicit company target mirrors the company",
			req:         dto.CreateReviewRequest{TargetType: "company", TargetID: testOtherCompanyID, Rating: 4, TransactionID: txTestTransactID},
			wantType:    domain.ReviewTargetCompany,
			wantTarget:  testOtherCompanyID,
			wantCompany: testOtherCompanyID,
		},
		{
			name:       "a user target carries no company",
			req:        dto.CreateReviewRequest{TargetType: "user", TargetID: testOtherID, Rating: 4, TransactionID: txTestTransactID},
			wantType:   domain.ReviewTargetUser,
			wantTarget: testOtherID,
		},
		{name: "unknown target type", req: dto.CreateReviewRequest{TargetType: "transaction", TargetID: farmerID, Rating: 4, TransactionID: txTestTransactID}, wantErr: domain.ErrInvalidReviewTargetType},
		{name: "self review", req: dto.CreateReviewRequest{TargetType: "user", TargetID: testUserID, Rating: 4, TransactionID: txTestTransactID}, wantErr: domain.ErrReviewTargetPartyMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := newFakeReviewRepo()
			uc := newTestReviewUC(reviewRepo)

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

func TestReviewUseCaseCreateReviewCompanyTarget(t *testing.T) {
	t.Parallel()

	unrelatedCompanyID := uuid.MustParse("12121212-1212-4212-8212-121212121212")

	t.Run("the counterparty's company is accepted", func(t *testing.T) {
		t.Parallel()

		reviewRepo := newFakeReviewRepo()
		uc := newTestReviewUC(reviewRepo)

		got, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testOtherCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.TargetID != testOtherCompanyID || got.CompanyID != testOtherCompanyID {
			t.Errorf("target/company = %v/%v, want %v", got.TargetID, got.CompanyID, testOtherCompanyID)
		}
		if len(reviewRepo.saved) != 1 {
			t.Fatalf("saved reviews = %d, want 1", len(reviewRepo.saved))
		}
	})

	t.Run("a company unrelated to the transaction is rejected", func(t *testing.T) {
		t.Parallel()

		reviewRepo := newFakeReviewRepo()
		uc := newTestReviewUC(reviewRepo)

		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: unrelatedCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrReviewTargetPartyMismatch) {
			t.Fatalf("error = %v, want %v", err, domain.ErrReviewTargetPartyMismatch)
		}
		if len(reviewRepo.saved) != 0 {
			t.Fatalf("saved reviews = %d, want 0", len(reviewRepo.saved))
		}
	})

	t.Run("the author's own company is rejected", func(t *testing.T) {
		t.Parallel()

		reviewRepo := newFakeReviewRepo()
		uc := newTestReviewUC(reviewRepo)

		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			CompanyID: testCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrReviewTargetPartyMismatch) {
			t.Fatalf("error = %v, want %v", err, domain.ErrReviewTargetPartyMismatch)
		}
		if len(reviewRepo.saved) != 0 {
			t.Fatalf("saved reviews = %d, want 0", len(reviewRepo.saved))
		}
	})

	t.Run("a counterparty without a company answers like a wrong target", func(t *testing.T) {
		t.Parallel()

		reviewRepo := newFakeReviewRepo()
		companyRepo := newFakeCompanyRepo()
		companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
			if ownerID == testOtherID {
				return nil, nil
			}
			return []domain.Company{*mustCompany()}, nil
		}
		uc := newTestReviewUCWithCompany(reviewRepo, companyRepo)

		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testOtherCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrReviewTargetPartyMismatch) {
			t.Fatalf("error = %v, want %v", err, domain.ErrReviewTargetPartyMismatch)
		}
	})

	t.Run("a company lookup failure propagates", func(t *testing.T) {
		t.Parallel()

		reviewRepo := newFakeReviewRepo()
		companyRepo := newFakeCompanyRepo()
		companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
			return nil, errFake
		}
		uc := newTestReviewUCWithCompany(reviewRepo, companyRepo)

		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testOtherCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, errFake) {
			t.Fatalf("error = %v, want %v", err, errFake)
		}
	})
}

func TestReviewUseCaseCreateReviewGuards(t *testing.T) {
	t.Parallel()

	t.Run("missing transaction id", func(t *testing.T) {
		t.Parallel()
		uc := newTestReviewUC(newFakeReviewRepo())
		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testCompanyID, Rating: 4,
		})
		if !errors.Is(err, domain.ErrTransactionRequired) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTransactionRequired)
		}
	})

	t.Run("transaction not completed", func(t *testing.T) {
		t.Parallel()
		reviewRepo := newFakeReviewRepo()
		txRepo := newFakeTxTransactionRepo()
		txRepo.stored.Status = domain.TransactionMatched
		matchRepo := newFakeTxMatchRepo()
		offerRepo := newFakeTxOfferRepo()
		offerRepo.stored.SupplierID = testOtherID
		requestRepo := newFakeTxRequestRepo()
		requestRepo.stored.BuyerID = testUserID
		uc := usecases.NewReviewUseCase(reviewRepo, txRepo, matchRepo, offerRepo, requestRepo, newFakeCompanyRepo(), newFakeTimer())
		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrTransactionNotCompleted) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTransactionNotCompleted)
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		t.Parallel()
		reviewRepo := newFakeReviewRepo()
		txRepo := newFakeTxTransactionRepo()
		txRepo.getByID = func(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
			return domain.Transaction{}, domain.ErrNotFound
		}
		matchRepo := newFakeTxMatchRepo()
		offerRepo := newFakeTxOfferRepo()
		offerRepo.stored.SupplierID = testOtherID
		requestRepo := newFakeTxRequestRepo()
		requestRepo.stored.BuyerID = testUserID
		uc := usecases.NewReviewUseCase(reviewRepo, txRepo, matchRepo, offerRepo, requestRepo, newFakeCompanyRepo(), newFakeTimer())
		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("principal is not a party", func(t *testing.T) {
		t.Parallel()
		uc := newTestReviewUC(newFakeReviewRepo())
		_, err := uc.CreateReview(principalCtxFor(uuid.New()), dto.CreateReviewRequest{
			TargetType: "company", TargetID: testCompanyID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
		}
	})

	t.Run("user target must be the other party", func(t *testing.T) {
		t.Parallel()
		uc := newTestReviewUC(newFakeReviewRepo())
		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "user", TargetID: testUserID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrReviewTargetPartyMismatch) {
			t.Fatalf("error = %v, want %v", err, domain.ErrReviewTargetPartyMismatch)
		}
	})

	t.Run("duplicate review for the same transaction", func(t *testing.T) {
		t.Parallel()
		reviewRepo := newFakeReviewRepo()
		reviewRepo.existsByTxAuth = func(ctx context.Context, transactionID, authorID uuid.UUID) (bool, error) {
			return true, nil
		}
		uc := newTestReviewUC(reviewRepo)
		_, err := uc.CreateReview(principalCtx(), dto.CreateReviewRequest{
			TargetType: "user", TargetID: testOtherID, Rating: 4, TransactionID: txTestTransactID,
		})
		if !errors.Is(err, domain.ErrReviewAlreadyExists) {
			t.Fatalf("error = %v, want %v", err, domain.ErrReviewAlreadyExists)
		}
		if len(reviewRepo.saved) != 0 {
			t.Fatalf("saved reviews = %d, want 0", len(reviewRepo.saved))
		}
	})
}

// TestReviewUseCaseGetAverageRating is the use-case half of the RF-15 aggregate:
// the count travels with the average, because 5.0 from two reviews and 5.0 from
// two hundred are not the same claim.
func TestReviewUseCaseGetAverageRating(t *testing.T) {
	t.Parallel()

	reviewRepo := newFakeReviewRepo()
	uc := newTestReviewUC(reviewRepo)

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
			uc := newTestReviewUC(reviewRepo)

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
			uc := newTestReviewUC(reviewRepo)

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
