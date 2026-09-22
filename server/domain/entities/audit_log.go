package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID
	ActorID    uuid.UUID
	Action     string
	TargetType string
	TargetID   uuid.UUID
	Metadata   json.RawMessage
	CreatedAt  time.Time
}

func NewAuditLog(actorID uuid.UUID, action string, targetType string, targetID uuid.UUID, metadata json.RawMessage, now time.Time) *AuditLog {
	return &AuditLog{
		ID:         uuid.New(),
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		CreatedAt:  now,
	}
}

const (
	AuditActionReportCreated   = "report_created"
	AuditActionReportApproved  = "report_approved"
	AuditActionReportRejected  = "report_rejected"
	AuditActionUserSuspended   = "user_suspended"
	AuditActionUserReactivated = "user_reactivated"
	AuditActionOfferingDeleted = "offering_deleted"
)
