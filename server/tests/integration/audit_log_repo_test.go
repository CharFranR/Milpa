package integration

import (
	"context"
	"encoding/json"
	"errors"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var testAuditLogIDs = []uuid.UUID{
	uuid.MustParse("55555555-5555-5555-5555-555555555501"),
	uuid.MustParse("55555555-5555-5555-5555-555555555502"),
	uuid.MustParse("55555555-5555-5555-5555-555555555503"),
	uuid.MustParse("55555555-5555-5555-5555-555555555504"),
	uuid.MustParse("55555555-5555-5555-5555-555555555505"),
}

var testAuditLogNotFoundID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555599")

var testAuditActorID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555511")
var testAuditSecondActorID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555512")

var testAuditReportTargetID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555521")
var testAuditUserTargetID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555522")
var testAuditOtherTargetID uuid.UUID = uuid.MustParse("55555555-5555-5555-5555-555555555523")

func newAuditLogFixture(id, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata string, createdAt time.Time) *domain.AuditLog {
	return &domain.AuditLog{
		ID:         id,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   json.RawMessage(metadata),
		CreatedAt:  createdAt,
	}
}

func setupAuditLogTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	userRepo := repository.NewUserRepository(TestPool)

	users := []*domain.User{
		{
			ID:           testAuditActorID,
			FirstName:    "Audit",
			LastName:     "Actor",
			Role:         domain.RoleAdmin,
			Email:        "audit-actor@example.com",
			PhoneNumber:  "4444-4444",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
		{
			ID:           testAuditSecondActorID,
			FirstName:    "Second",
			LastName:     "Actor",
			Role:         domain.RoleAdmin,
			Email:        "audit-actor-2@example.com",
			PhoneNumber:  "5555-5555",
			PasswordHash: "hash",
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}

	for _, u := range users {
		if _, err := userRepo.Save(context.Background(), u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}
}

// assertMetadataEqual compares metadata semantically: PostgreSQL jsonb does
// not preserve key order or whitespace, so a raw byte comparison would be
// flaky even for identical payloads.
func assertMetadataEqual(t *testing.T, want, got json.RawMessage) {
	t.Helper()

	var wantMap, gotMap map[string]any
	if err := json.Unmarshal(want, &wantMap); err != nil {
		t.Fatalf("invalid expected metadata %s: %v", want, err)
	}
	if err := json.Unmarshal(got, &gotMap); err != nil {
		t.Fatalf("invalid stored metadata %s: %v", got, err)
	}

	if !reflect.DeepEqual(wantMap, gotMap) {
		t.Errorf("metadata = %s, want %s", got, want)
	}
}

func TestAuditLogSaveAndFindByID(t *testing.T) {
	setupAuditLogTestData(t)
	db := repository.NewAuditLogRepository(TestPool)

	metadata := `{"report_id":"88888888-8888-8888-8888-888888888801","outcome":"approved"}`
	saved := newAuditLogFixture(testAuditLogIDs[0], testAuditActorID, domain.AuditActionReportApproved, "report", testAuditReportTargetID, metadata, fixedTime)
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
			ID:          testAuditLogIDs[0],
			ExpectedErr: nil,
		},
		{
			// FINDING (production behaviour, not changed here): unlike the
			// other repositories, AuditLogRepositoryImpl.FindByID does not
			// translate pgx.ErrNoRows into domain.ErrNotFound, so callers
			// must compare against the pgx sentinel instead.
			Name:        "Audit Log Not Found",
			ID:          testAuditLogNotFoundID,
			ExpectedErr: pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			entry, err := db.FindByID(context.Background(), tt.ID)

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

			if entry == nil {
				t.Fatal("FindByID() returned nil audit log")
			}
			if entry.ID != saved.ID {
				t.Errorf("FindByID() ID = %v, want %v", entry.ID, saved.ID)
			}
			if entry.ActorID != saved.ActorID {
				t.Errorf("FindByID() ActorID = %v, want %v", entry.ActorID, saved.ActorID)
			}
			if entry.Action != saved.Action {
				t.Errorf("FindByID() Action = %v, want %v", entry.Action, saved.Action)
			}
			if entry.TargetType != saved.TargetType {
				t.Errorf("FindByID() TargetType = %v, want %v", entry.TargetType, saved.TargetType)
			}
			if entry.TargetID != saved.TargetID {
				t.Errorf("FindByID() TargetID = %v, want %v", entry.TargetID, saved.TargetID)
			}
			if !entry.CreatedAt.Equal(saved.CreatedAt) {
				t.Errorf("FindByID() CreatedAt = %v, want %v", entry.CreatedAt, saved.CreatedAt)
			}
			assertMetadataEqual(t, saved.Metadata, entry.Metadata)
		})
	}
}

func TestAuditLogFindAllOrderingAndPagination(t *testing.T) {
	setupAuditLogTestData(t)
	db := repository.NewAuditLogRepository(TestPool)

	// Distinct created_at values keep ORDER BY created_at DESC deterministic.
	saved := make([]*domain.AuditLog, 0, len(testAuditLogIDs))
	for i, id := range testAuditLogIDs {
		entry := newAuditLogFixture(id, testAuditActorID, domain.AuditActionReportCreated, "report", testAuditReportTargetID,
			`{"report_id":"88888888-8888-8888-8888-888888888801"}`, fixedTime.Add(time.Duration(i)*time.Minute))
		if err := db.Save(context.Background(), entry); err != nil {
			t.Fatalf("Save() audit log %d: %v", i, err)
		}
		saved = append(saved, entry)
	}

	tests := []struct {
		Name          string
		Page          int
		PageSize      int
		ExpectedLen   int
		ExpectedTotal int
		ExpectedIDs   []uuid.UUID
	}{
		{
			Name:          "First page holds the newest entries",
			Page:          1,
			PageSize:      2,
			ExpectedLen:   2,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[4].ID, saved[3].ID},
		},
		{
			Name:          "Second page holds the middle entries",
			Page:          2,
			PageSize:      2,
			ExpectedLen:   2,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[2].ID, saved[1].ID},
		},
		{
			Name:          "Last page holds the oldest entry only",
			Page:          3,
			PageSize:      2,
			ExpectedLen:   1,
			ExpectedTotal: 5,
			ExpectedIDs:   []uuid.UUID{saved[0].ID},
		},
		{
			Name:          "Page beyond the last one is empty but keeps the total",
			Page:          4,
			PageSize:      2,
			ExpectedLen:   0,
			ExpectedTotal: 5,
			ExpectedIDs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			logs, total, err := db.FindAll(context.Background(), "", "", "", tt.Page, tt.PageSize)
			if err != nil {
				t.Errorf("FindAll() unexpected error: %v", err)
				return
			}

			if total != tt.ExpectedTotal {
				t.Errorf("FindAll() total = %d, want %d", total, tt.ExpectedTotal)
			}
			if len(logs) != tt.ExpectedLen {
				t.Fatalf("FindAll() got %d entries, want %d", len(logs), tt.ExpectedLen)
			}

			for i, wantID := range tt.ExpectedIDs {
				if logs[i].ID != wantID {
					t.Errorf("FindAll()[%d].ID = %v, want %v", i, logs[i].ID, wantID)
				}
				assertMetadataEqual(t, saved[0].Metadata, logs[i].Metadata)
			}
		})
	}
}

func TestAuditLogFindAllFilters(t *testing.T) {
	setupAuditLogTestData(t)
	db := repository.NewAuditLogRepository(TestPool)

	fixtures := []*domain.AuditLog{
		newAuditLogFixture(testAuditLogIDs[0], testAuditActorID, domain.AuditActionReportCreated, "report", testAuditReportTargetID, `{"report_id":"r0"}`, fixedTime),
		newAuditLogFixture(testAuditLogIDs[1], testAuditActorID, domain.AuditActionReportApproved, "report", testAuditReportTargetID, `{"report_id":"r1"}`, fixedTime.Add(time.Minute)),
		newAuditLogFixture(testAuditLogIDs[2], testAuditSecondActorID, domain.AuditActionUserSuspended, "user", testAuditUserTargetID, `{"user_id":"u2"}`, fixedTime.Add(2*time.Minute)),
		newAuditLogFixture(testAuditLogIDs[3], testAuditSecondActorID, domain.AuditActionReportCreated, "report", testAuditOtherTargetID, `{"report_id":"r3"}`, fixedTime.Add(3*time.Minute)),
	}

	for _, entry := range fixtures {
		if err := db.Save(context.Background(), entry); err != nil {
			t.Fatalf("Save() audit log %s: %v", entry.ID, err)
		}
	}

	tests := []struct {
		Name          string
		Action        string
		ActorID       string
		TargetType    string
		ExpectedTotal int
	}{
		{
			Name:          "No filters returns every entry",
			ExpectedTotal: 4,
		},
		{
			Name:          "Action report_created",
			Action:        domain.AuditActionReportCreated,
			ExpectedTotal: 2,
		},
		{
			Name:          "Action user_suspended",
			Action:        domain.AuditActionUserSuspended,
			ExpectedTotal: 1,
		},
		{
			Name:          "Actor filter uses the string form of the uuid",
			ActorID:       testAuditActorID.String(),
			ExpectedTotal: 2,
		},
		{
			Name:          "Second actor filter",
			ActorID:       testAuditSecondActorID.String(),
			ExpectedTotal: 2,
		},
		{
			Name:          "Target type user",
			TargetType:    "user",
			ExpectedTotal: 1,
		},
		{
			Name:          "Target type report",
			TargetType:    "report",
			ExpectedTotal: 3,
		},
		{
			Name:          "Action and actor combined",
			Action:        domain.AuditActionReportCreated,
			ActorID:       testAuditActorID.String(),
			ExpectedTotal: 1,
		},
		{
			Name:          "Action and target type with no match",
			Action:        domain.AuditActionReportCreated,
			TargetType:    "user",
			ExpectedTotal: 0,
		},
		{
			Name:          "Unknown action is a plain varchar filter and matches nothing",
			Action:        "does_not_exist",
			ExpectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			logs, total, err := db.FindAll(context.Background(), tt.Action, tt.ActorID, tt.TargetType, 1, 10)
			if err != nil {
				t.Errorf("FindAll() unexpected error: %v", err)
				return
			}

			if total != tt.ExpectedTotal {
				t.Errorf("FindAll() total = %d, want %d", total, tt.ExpectedTotal)
			}
			if len(logs) != tt.ExpectedTotal {
				t.Errorf("FindAll() got %d entries, want %d", len(logs), tt.ExpectedTotal)
			}
		})
	}
}
