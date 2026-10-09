package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func interestFixture(liquidationID, buyerID uuid.UUID, at time.Time) domain.LiquidationInterest {
	return *domain.NewLiquidationInterest(liquidationID, buyerID, at)
}

func TestLiquidationExpressInterest(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("Wholesale corn", "wholesale")
	closed := liquidationFixture("Closed corn", "wholesale")
	if err := closed.Close(fixedTime); err != nil {
		t.Fatalf("Close: %v", err)
	}

	tests := []struct {
		name           string
		ctx            context.Context
		liquidation    domain.Liquidation
		findVisibleErr error
		wantErr        error
		wantSaved      bool
		wantRole       domain.RoleOptions
	}{
		{name: "a minorista can express interest", ctx: principalCtx(), liquidation: open, wantSaved: true, wantRole: domain.RoleCompradorMinorista},
		{name: "a mayorista detallista can express interest", ctx: mayoristaCtx(), liquidation: open, wantSaved: true, wantRole: domain.RoleCompradorMayoristaDetallista},
		{name: "a mayorista corporativo can express interest", ctx: corporateCtx(), liquidation: open, wantSaved: true, wantRole: domain.RoleCompradorMayoristaCorporativo},
		{name: "anonymous is refused", ctx: context.Background(), liquidation: open, wantErr: auth.ErrUnauthenticated},
		{name: "a farmer is refused", ctx: farmerCtx(), liquidation: open, wantErr: domain.ErrForbidden},
		{name: "an invisible liquidation answers not found", ctx: principalCtx(), liquidation: open, findVisibleErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "a non-open liquidation is a conflict", ctx: principalCtx(), liquidation: closed, wantErr: domain.ErrLiquidationNotOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeLiquidationRepo(tt.liquidation)
			repo.findVisibleErr = tt.findVisibleErr
			uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

			err := uc.ExpressInterest(tt.ctx, tt.liquidation.ID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ExpressInterest() error = %v, want %v", err, tt.wantErr)
				}
				if len(repo.interests) != 0 {
					t.Fatalf("a refused caller saved %d interests, want none", len(repo.interests))
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpressInterest() error: %v", err)
			}
			if !tt.wantSaved || len(repo.interests) != 1 {
				t.Fatalf("saved %d interests, want 1", len(repo.interests))
			}
			if repo.interests[0].BuyerID != testUserID {
				t.Errorf("buyer_id = %v, want %v", repo.interests[0].BuyerID, testUserID)
			}
			if repo.interests[0].LiquidationID != tt.liquidation.ID {
				t.Errorf("liquidation_id = %v, want %v", repo.interests[0].LiquidationID, tt.liquidation.ID)
			}
			if !repo.viewerSeen || repo.lastViewer.Role != tt.wantRole {
				t.Errorf("viewer role = %v (seen=%v), want %v", repo.lastViewer.Role, repo.viewerSeen, tt.wantRole)
			}
		})
	}
}

func TestLiquidationExpressInterestRejectsADuplicate(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("Wholesale corn", "wholesale")
	existing := interestFixture(open.ID, testUserID, fixedTime)

	repo := newFakeLiquidationRepo(open)
	repo.interests = append(repo.interests, existing)
	uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

	err := uc.ExpressInterest(principalCtx(), open.ID)
	if !errors.Is(err, domain.ErrInterestAlreadyExists) {
		t.Fatalf("ExpressInterest() error = %v, want %v", err, domain.ErrInterestAlreadyExists)
	}
	if len(repo.interests) != 1 {
		t.Errorf("interests = %d, want the original one", len(repo.interests))
	}
}

func TestLiquidationListInterests(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("Wholesale corn", "wholesale")
	other := interestFixture(open.ID, testUserID, fixedTime)

	repo := newFakeLiquidationRepo(open)
	repo.interests = append(repo.interests, other)
	uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

	got, err := uc.ListInterests(farmerCtxFor(testOtherID), open.ID)
	if err != nil {
		t.Fatalf("ListInterests() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListInterests() = %d, want 1", len(got))
	}
	if got[0].ID != other.ID || got[0].LiquidationID != open.ID || got[0].BuyerID != testUserID {
		t.Errorf("interest = %+v, want id=%v liquidation=%v buyer=%v", got[0], other.ID, open.ID, testUserID)
	}
	if got[0].BuyerName != "John Doe" {
		t.Errorf("buyer_name = %q, want %q", got[0].BuyerName, "John Doe")
	}
	if !got[0].CreatedAt.Equal(fixedTime) {
		t.Errorf("created_at = %v, want %v", got[0].CreatedAt, fixedTime)
	}

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{name: "a non-supplier is refused", ctx: principalCtx(), wantErr: domain.ErrForbidden},
		{name: "another supplier is refused", ctx: farmerCtx(), wantErr: domain.ErrForbidden},
		{name: "anonymous is refused", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := uc.ListInterests(tt.ctx, open.ID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListInterests() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidationAssignManual(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("Manual corn", "public")
	interested := interestFixture(open.ID, testUserID, fixedTime)
	notInterested := uuid.MustParse("12121212-1212-1212-1212-121212121212")

	closed := liquidationFixture("Closed corn", "public")
	if err := closed.Close(fixedTime); err != nil {
		t.Fatalf("Close: %v", err)
	}

	assigned := liquidationFixture("Assigned corn", "public")
	if err := assigned.Assign(testUserID, fixedTime); err != nil {
		t.Fatalf("Assign: %v", err)
	}

	tests := []struct {
		name        string
		ctx         context.Context
		liquidation domain.Liquidation
		buyerID     *uuid.UUID
		seed        []domain.LiquidationInterest
		wantErr     error
		wantBuyer   uuid.UUID
	}{
		{
			name:        "the supplier assigns an interested buyer",
			ctx:         farmerCtxFor(testOtherID),
			liquidation: open,
			buyerID:     &testUserID,
			seed:        []domain.LiquidationInterest{interested},
			wantBuyer:   testUserID,
		},
		{
			name:        "a buyer that never expressed interest is refused",
			ctx:         farmerCtxFor(testOtherID),
			liquidation: open,
			buyerID:     &notInterested,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrBuyerDidNotExpressInterest,
		},
		{
			name:        "a missing buyer_id is refused",
			ctx:         farmerCtxFor(testOtherID),
			liquidation: open,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrBuyerDidNotExpressInterest,
		},
		{
			name:        "a non-supplier is refused",
			ctx:         farmerCtx(),
			liquidation: open,
			buyerID:     &testUserID,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrForbidden,
		},
		{
			name:        "a buyer is refused",
			ctx:         principalCtx(),
			liquidation: open,
			buyerID:     &testUserID,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrForbidden,
		},
		{
			name:        "a closed liquidation is a conflict",
			ctx:         farmerCtxFor(testOtherID),
			liquidation: closed,
			buyerID:     &testUserID,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrLiquidationCannotAssign,
		},
		{
			name:        "an assigned liquidation is a conflict",
			ctx:         farmerCtxFor(testOtherID),
			liquidation: assigned,
			buyerID:     &testUserID,
			seed:        []domain.LiquidationInterest{interested},
			wantErr:     domain.ErrLiquidationCannotAssign,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeLiquidationRepo(tt.liquidation)
			repo.interests = append(repo.interests, tt.seed...)
			uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

			err := uc.AssignLiquidation(tt.ctx, tt.liquidation.ID, tt.buyerID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("AssignLiquidation() error = %v, want %v", err, tt.wantErr)
				}
				if len(repo.updated) != 0 {
					t.Fatalf("a refused assignment wrote %d updates, want none", len(repo.updated))
				}
				return
			}
			if err != nil {
				t.Fatalf("AssignLiquidation() error: %v", err)
			}
			if len(repo.updated) != 1 {
				t.Fatalf("updates = %d, want 1", len(repo.updated))
			}
			updated := repo.updated[0]
			if updated.Status != domain.LiquidationAssigned {
				t.Errorf("status = %v, want %v", updated.Status, domain.LiquidationAssigned)
			}
			if updated.AssignedBuyerID == nil || *updated.AssignedBuyerID != tt.wantBuyer {
				t.Errorf("assigned_buyer_id = %v, want %v", updated.AssignedBuyerID, tt.wantBuyer)
			}
			if updated.ClosedAt == nil || !updated.ClosedAt.Equal(fixedTime) {
				t.Errorf("closed_at = %v, want %v", updated.ClosedAt, fixedTime)
			}
		})
	}
}

func TestLiquidationAssignFirstComePicksTheEarliestInterest(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("First come corn", "public")
	open.AllocationMethod = domain.AllocationFirstCome

	buyerEarly := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	buyerLate := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	buyerTieA := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	buyerTieB := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	interestLower := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	interestHigher := uuid.MustParse("00000000-0000-0000-0000-00000000000b")

	first := interestFixture(open.ID, buyerEarly, fixedTime.Add(-time.Hour))
	second := interestFixture(open.ID, buyerLate, fixedTime)
	tieLower := interestFixture(open.ID, buyerTieA, fixedTime.Add(-2*time.Hour))
	tieLower.ID = interestLower
	tieHigher := interestFixture(open.ID, buyerTieB, fixedTime.Add(-2*time.Hour))
	tieHigher.ID = interestHigher

	repo := newFakeLiquidationRepo(open)
	repo.interests = append(repo.interests, second, first, tieHigher, tieLower)
	uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

	passed := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	if err := uc.AssignLiquidation(farmerCtxFor(testOtherID), open.ID, &passed); err != nil {
		t.Fatalf("AssignLiquidation() error: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updates = %d, want 1", len(repo.updated))
	}
	got := repo.updated[0].AssignedBuyerID
	if got == nil || *got != buyerTieA {
		t.Fatalf("assigned_buyer_id = %v, want the earliest tie-broken by interest id (%v)", got, buyerTieA)
	}
}

func TestLiquidationAssignFirstComeWithNoInterest(t *testing.T) {
	t.Parallel()

	open := liquidationFixture("First come corn", "public")
	open.AllocationMethod = domain.AllocationFirstCome

	repo := newFakeLiquidationRepo(open)
	uc := usecases.NewLiquidationUseCase(repo, newFakeUserRepo(), newFakeTimer())

	err := uc.AssignLiquidation(farmerCtxFor(testOtherID), open.ID, nil)
	if !errors.Is(err, domain.ErrNoInterestToAssign) {
		t.Fatalf("AssignLiquidation() error = %v, want %v", err, domain.ErrNoInterestToAssign)
	}
	if len(repo.updated) != 0 {
		t.Errorf("updates = %d, want none", len(repo.updated))
	}
}
