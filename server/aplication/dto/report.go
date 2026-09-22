package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateReportRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
}

type ResolveReportRequest struct {
	Action string `json:"action"`
}

type SuspendUserRequest struct {
	Action string `json:"action"`
}

type ReportResponse struct {
	ID         uuid.UUID  `json:"id"`
	Reporter   UserBrief  `json:"reporter"`
	TargetType string     `json:"target_type"`
	Target     TargetBrief `json:"target"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	ResolvedBy *uuid.UUID `json:"resolved_by,omitempty"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type UserBrief struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type TargetBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AuditLogResponse struct {
	ID         uuid.UUID `json:"id"`
	ActorID    uuid.UUID `json:"actor_id"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Metadata   any       `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type PaginatedReportsResponse struct {
	Items []ReportResponse `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

type PaginatedAuditLogsResponse struct {
	Items []AuditLogResponse `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}
