package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReportTargetType string

const (
	ReportTargetOffering ReportTargetType = "offering"
	ReportTargetUser     ReportTargetType = "user"
)

type ReportStatus string

const (
	ReportPending  ReportStatus = "pending"
	ReportApproved ReportStatus = "approved"
	ReportRejected ReportStatus = "rejected"
)

type Report struct {
	ID         uuid.UUID
	ReporterID uuid.UUID
	TargetType ReportTargetType
	TargetID   uuid.UUID
	Reason     string
	Status     ReportStatus
	ResolvedBy *uuid.UUID
	ResolvedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewReport(reporterID uuid.UUID, targetType ReportTargetType, targetID uuid.UUID, reason string, now time.Time) (*Report, error) {
	if reporterID == uuid.Nil {
		return nil, ErrReporterRequired
	}
	if targetType != ReportTargetOffering && targetType != ReportTargetUser {
		return nil, ErrInvalidReportTargetType
	}
	if targetID == uuid.Nil {
		return nil, ErrTargetRequired
	}
	if reason == "" {
		return nil, ErrReasonRequired
	}
	if reporterID == targetID && targetType == ReportTargetUser {
		return nil, ErrSelfReport
	}

	return &Report{
		ID:         uuid.New(),
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
		Status:     ReportPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (r *Report) Approve(adminID uuid.UUID, now time.Time) {
	r.Status = ReportApproved
	r.ResolvedBy = &adminID
	r.ResolvedAt = &now
	r.UpdatedAt = now
}

func (r *Report) Reject(adminID uuid.UUID, now time.Time) {
	r.Status = ReportRejected
	r.ResolvedBy = &adminID
	r.ResolvedAt = &now
	r.UpdatedAt = now
}

func (r Report) IsPending() bool {
	return r.Status == ReportPending
}

func (r Report) IsResolved() bool {
	return r.Status == ReportApproved || r.Status == ReportRejected
}
