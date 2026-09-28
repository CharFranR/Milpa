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

// TestOfferingUseCaseDeleteOfferingOwnership is the IDOR on the delete path.
// The route is mounted behind authentication only, so before the ownership
// check any valid token could destroy any farmer's product.
func TestOfferingUseCaseDeleteOfferingOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		repoErr     error
		deleteErr   error
		wantErr     error
		wantDeleted bool
	}{
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "foreign authenticated user", ctx: principalCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "owner", ctx: principalCtx(), wantDeleted: true},
		{name: "admin", ctx: reportAdminCtx(), wantDeleted: true},
		{name: "offering not found", ctx: principalCtx(), repoErr: domain.ErrNotFound, wantErr: domain.ErrNotFound},
		{name: "repo error", ctx: principalCtx(), repoErr: errFake, wantErr: errFake},
		{name: "delete error", ctx: principalCtx(), deleteErr: errFake, wantErr: errFake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			if tt.repoErr != nil {
				offeringRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
					return nil, tt.repoErr
				}
			}
			if tt.deleteErr != nil {
				offeringRepo.delete = func(ctx context.Context, id uuid.UUID) error {
					return tt.deleteErr
				}
			}

			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			err := uc.DeleteOffering(tt.ctx, testOfferingID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantDeleted {
				if len(offeringRepo.deleted) != 1 || offeringRepo.deleted[0] != testOfferingID {
					t.Errorf("deleted = %v, want [%v]", offeringRepo.deleted, testOfferingID)
				}
				return
			}
			if len(offeringRepo.deleted) != 0 {
				t.Errorf("a refused caller still deleted %v, want nothing", offeringRepo.deleted)
			}
		})
	}
}

// TestOfferingUseCaseUpdateOfferingOwnership is the same IDOR on the update
// path, spelled out separately because the two are separate routes.
func TestOfferingUseCaseUpdateOfferingOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ctx      context.Context
		wantErr  error
		wantName string
	}{
		{name: "unauthenticated", ctx: context.Background(), wantErr: auth.ErrUnauthenticated},
		{name: "foreign authenticated user", ctx: principalCtxFor(testOtherID), wantErr: domain.ErrForbidden},
		{name: "owner", ctx: principalCtx(), wantName: "Renamed by owner"},
		{name: "admin", ctx: reportAdminCtx(), wantName: "Renamed by admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offeringRepo := newFakeOfferingRepo()
			uc := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})

			err := uc.UpdateOffering(tt.ctx, testOfferingID, dto.UpdateOfferingRequest{Name: strPtr("Renamed")})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if len(offeringRepo.updated) != 0 {
					t.Errorf("a refused caller still wrote %d offerings, want 0", len(offeringRepo.updated))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(offeringRepo.updated) != 1 {
				t.Fatalf("updated offerings = %d, want 1", len(offeringRepo.updated))
			}
		})
	}
}

// TestModerationDeleteOfferingStillReachesAdminModeration guards the carve-out:
// the admin moderation route delegates to the same use case, so an ownership
// check that also refused admins would silently break moderation.
func TestModerationDeleteOfferingStillReachesAdminModeration(t *testing.T) {
	t.Parallel()

	offeringRepo := newFakeOfferingRepo()
	offeringUC := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})
	auditRepo := newFakeAuditLogRepo()
	moderation := usecases.NewModerationUseCase(newFakeUserRepo(), offeringRepo, auditRepo, offeringUC, newFakeTimer())

	err := moderation.DeleteOffering(reportAdminCtx(), testOfferingID)
	if err != nil {
		t.Fatalf("admin moderation delete: %v", err)
	}
	if len(offeringRepo.deleted) != 1 {
		t.Errorf("moderation deleted %v, want the offering to be deleted", offeringRepo.deleted)
	}
}

func TestModerationDeleteOfferingRefusesNonAdmin(t *testing.T) {
	t.Parallel()

	offeringRepo := newFakeOfferingRepo()
	offeringUC := usecases.NewOfferingUseCase(offeringRepo, newFakeUserRepo(), newFakeTimer(), &fakeFuzzyRetrival{}, &fakeInvalidator{})
	auditRepo := newFakeAuditLogRepo()
	moderation := usecases.NewModerationUseCase(newFakeUserRepo(), offeringRepo, auditRepo, offeringUC, newFakeTimer())

	err := moderation.DeleteOffering(principalCtx(), testOfferingID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
	}
	if len(offeringRepo.deleted) != 0 {
		t.Errorf("a non-admin deleted %v, want nothing", offeringRepo.deleted)
	}
}
