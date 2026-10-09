package usecases_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func TestModerationListUsersReturnsTheAdminProjection(t *testing.T) {
	t.Parallel()

	suspendedAt := fixedTime.Add(48 * time.Hour)
	user := mustUser()
	user.Role = domain.RoleAgricultor
	user.PhoneNumber = "555-1234"
	user.PhotoURL = "http://images.milpa.com/john.png"
	user.SuspendedAt = &suspendedAt

	userRepo := newFakeUserRepo()
	userRepo.list = func(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
		return []domain.User{*user}, 1, nil
	}
	uc := usecases.NewModerationUseCase(userRepo, newFakeOfferingRepo(), newFakeAuditLogRepo(), newFakeTimer())

	result, err := uc.ListUsers(reportAdminCtx(), 0, 0)
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if result.Total != 1 || result.Page != 1 || result.Size != 20 {
		t.Fatalf("pagination = total %d page %d size %d, want 1/1/20", result.Total, result.Page, result.Size)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	item := result.Items[0]
	if item.ID != testUserID || item.FirstName != "John" || item.LastName != "Doe" || item.Email != "user@milpa.com.ni" {
		t.Errorf("identity = %+v, want the mapped user", item)
	}
	if item.PhoneNumber != "555-1234" || item.Role != domain.RoleAgricultor || item.Department != "Leon" || item.Municipality != "Leon" || item.PhotoURL != "http://images.milpa.com/john.png" {
		t.Errorf("profile = %+v, want the mapped fields", item)
	}
	if item.SuspendedAt == nil || !item.SuspendedAt.Equal(suspendedAt) {
		t.Errorf("suspended_at = %v, want %v", item.SuspendedAt, suspendedAt)
	}
	if !item.CreatedAt.Equal(fixedTime) || !item.UpdatedAt.Equal(fixedTime) {
		t.Errorf("timestamps = %v/%v, want %v", item.CreatedAt, item.UpdatedAt, fixedTime)
	}
}

func TestModerationListUsersReportsAnActiveUserAsUnsuspended(t *testing.T) {
	t.Parallel()

	user := mustUser()
	user.Role = domain.RoleCompradorMinorista

	userRepo := newFakeUserRepo()
	userRepo.list = func(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
		return []domain.User{*user}, 1, nil
	}
	uc := usecases.NewModerationUseCase(userRepo, newFakeOfferingRepo(), newFakeAuditLogRepo(), newFakeTimer())

	result, err := uc.ListUsers(reportAdminCtx(), 1, 20)
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if result.Items[0].SuspendedAt != nil {
		t.Errorf("suspended_at = %v, want nil for an active user", result.Items[0].SuspendedAt)
	}
}

func TestModerationListAuditLogsReturnsTheAdminProjection(t *testing.T) {
	t.Parallel()

	logID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	metadata := json.RawMessage(`{"user_id":"` + testOtherID.String() + `"}`)

	auditRepo := newFakeAuditLogRepo()
	auditRepo.findAll = func(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error) {
		return []domain.AuditLog{{
			ID:         logID,
			ActorID:    testAdminID,
			Action:     domain.AuditActionUserSuspended,
			TargetType: "user",
			TargetID:   testOtherID,
			Metadata:   metadata,
			CreatedAt:  fixedTime,
		}}, 1, nil
	}
	uc := usecases.NewModerationUseCase(newFakeUserRepo(), newFakeOfferingRepo(), auditRepo, newFakeTimer())

	result, err := uc.ListAuditLogs(reportAdminCtx(), "", "", "", 2, 5)
	if err != nil {
		t.Fatalf("ListAuditLogs() error: %v", err)
	}
	if result.Total != 1 || result.Page != 2 || result.Size != 5 {
		t.Fatalf("pagination = total %d page %d size %d, want 1/2/5", result.Total, result.Page, result.Size)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	item := result.Items[0]
	if item.ID != logID || item.ActorID != testAdminID || item.Action != domain.AuditActionUserSuspended {
		t.Errorf("identity = %+v, want the mapped audit row", item)
	}
	if item.TargetType != "user" || item.TargetID != testOtherID || !item.CreatedAt.Equal(fixedTime) {
		t.Errorf("target = %+v, want the mapped user target", item)
	}
	decoded, ok := item.Metadata.(map[string]interface{})
	if !ok || decoded["user_id"] != testOtherID.String() {
		t.Errorf("metadata = %#v, want the decoded user_id", item.Metadata)
	}
}

func TestModerationSetUserRolePersistsTheChange(t *testing.T) {
	t.Parallel()

	for _, role := range []domain.RoleOptions{domain.RoleAuditor, domain.RoleAgricultor} {
		t.Run(role.String(), func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			auditRepo := newFakeAuditLogRepo()
			uc := usecases.NewModerationUseCase(userRepo, newFakeOfferingRepo(), auditRepo, newFakeTimer())

			if err := uc.SetUserRole(reportAdminCtx(), testOtherID, dto.SetUserRoleRequest{Role: role}); err != nil {
				t.Fatalf("SetUserRole(%s) error: %v", role, err)
			}
			if len(userRepo.updated) != 1 {
				t.Fatalf("updated = %d users, want 1", len(userRepo.updated))
			}
			if userRepo.updated[0].ID != testOtherID || userRepo.updated[0].Role != role {
				t.Errorf("persisted = %+v, want %v with role %s", userRepo.updated[0], testOtherID, role)
			}
			if len(auditRepo.saved) != 1 || auditRepo.saved[0].Action != domain.AuditActionUserRoleUpdated {
				t.Errorf("audit = %+v, want one user_role_updated entry", auditRepo.saved)
			}
		})
	}
}

func TestModerationSetUserRoleRejectsAnUnknownRole(t *testing.T) {
	t.Parallel()

	userRepo := newFakeUserRepo()
	userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		t.Error("FindByID must not be called for an unknown role")
		return nil, nil
	}
	auditRepo := newFakeAuditLogRepo()
	uc := usecases.NewModerationUseCase(userRepo, newFakeOfferingRepo(), auditRepo, newFakeTimer())

	err := uc.SetUserRole(reportAdminCtx(), testOtherID, dto.SetUserRoleRequest{Role: domain.RoleOptions(99)})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("SetUserRole() error = %v, want ErrInvalidInput", err)
	}
	if len(userRepo.updated) != 0 || len(auditRepo.saved) != 0 {
		t.Errorf("an unknown role wrote %d users and %d audit logs, want 0", len(userRepo.updated), len(auditRepo.saved))
	}
}

func TestModerationAdminActionsRefuseANonAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(uc *usecases.ModerationUseCaseImpl) error
	}{
		{
			name: "list users",
			call: func(uc *usecases.ModerationUseCaseImpl) error {
				_, err := uc.ListUsers(principalCtx(), 1, 20)
				return err
			},
		},
		{
			name: "set role",
			call: func(uc *usecases.ModerationUseCaseImpl) error {
				return uc.SetUserRole(principalCtx(), testOtherID, dto.SetUserRoleRequest{Role: domain.RoleAuditor})
			},
		},
		{
			name: "list audit logs",
			call: func(uc *usecases.ModerationUseCaseImpl) error {
				_, err := uc.ListAuditLogs(principalCtx(), "", "", "", 1, 20)
				return err
			},
		},
		{
			name: "suspend user",
			call: func(uc *usecases.ModerationUseCaseImpl) error {
				return uc.SuspendUser(principalCtx(), testOtherID, dto.SuspendUserRequest{Action: "suspend"})
			},
		},
		{
			name: "delete offering",
			call: func(uc *usecases.ModerationUseCaseImpl) error {
				return uc.DeleteOffering(principalCtx(), testOfferingID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			userRepo.list = func(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
				t.Error("a refused caller must not list users")
				return nil, 0, nil
			}
			userRepo.update = func(ctx context.Context, user *domain.User) error {
				t.Error("a refused caller must not persist a user")
				return nil
			}
			offeringRepo := newFakeOfferingRepo()
			offeringRepo.delete = func(ctx context.Context, id uuid.UUID) error {
				t.Error("a refused caller must not delete an offering")
				return nil
			}
			auditRepo := newFakeAuditLogRepo()
			auditRepo.save = func(ctx context.Context, log *domain.AuditLog) error {
				t.Error("a refused caller must not write an audit log")
				return nil
			}
			uc := usecases.NewModerationUseCase(userRepo, offeringRepo, auditRepo, newFakeTimer())

			if err := tt.call(uc); !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("error = %v, want ErrForbidden", err)
			}
			if len(userRepo.updated) != 0 || len(offeringRepo.deleted) != 0 || len(auditRepo.saved) != 0 {
				t.Errorf("a refused caller wrote %d users, %d offerings and %d audit logs, want 0", len(userRepo.updated), len(offeringRepo.deleted), len(auditRepo.saved))
			}
		})
	}
}

func TestModerationAdminActionsRequireAuthentication(t *testing.T) {
	t.Parallel()

	uc := usecases.NewModerationUseCase(newFakeUserRepo(), newFakeOfferingRepo(), newFakeAuditLogRepo(), newFakeTimer())

	if _, err := uc.ListUsers(context.Background(), 1, 20); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("ListUsers() error = %v, want ErrUnauthenticated", err)
	}
	if _, err := uc.ListAuditLogs(context.Background(), "", "", "", 1, 20); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("ListAuditLogs() error = %v, want ErrUnauthenticated", err)
	}
	if err := uc.SetUserRole(context.Background(), testOtherID, dto.SetUserRoleRequest{Role: domain.RoleAuditor}); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("SetUserRole() error = %v, want ErrUnauthenticated", err)
	}
	if err := uc.SuspendUser(context.Background(), testOtherID, dto.SuspendUserRequest{Action: "suspend"}); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("SuspendUser() error = %v, want ErrUnauthenticated", err)
	}
	if err := uc.DeleteOffering(context.Background(), testOfferingID); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("DeleteOffering() error = %v, want ErrUnauthenticated", err)
	}
}
