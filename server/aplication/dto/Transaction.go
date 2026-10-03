package dto

import (
	"time"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type TransactionDTO struct {
	ID                          *uuid.UUID                   `json:"id"`
	MatchID                     *uuid.UUID                   `json:"match_id"`
	Status                      domain.TransactionStatus     `json:"status"`
	BuyerStartConfirmedAt       *time.Time                   `json:"buyer_start_confirmed_at"`
	SupplierStartConfirmedAt    *time.Time                   `json:"supplier_start_confirmed_at"`
	BuyerDeliveryConfirmedAt    *time.Time                   `json:"buyer_delivery_confirmed_at"`
	SupplierDeliveryConfirmedAt *time.Time                   `json:"supplier_delivery_confirmed_at"`
	CancelledBy                 *uuid.UUID                   `json:"cancelled_by"`
	CancelReason                string                       `json:"cancel_reason"`
	CreatedAt                   time.Time                    `json:"created_at"`
	UpdatedAt                   time.Time                    `json:"updated_at"`
	History                     []TransactionHistoryEntryDTO `json:"history"`
}

type TransactionHistoryEntryDTO struct {
	Status       domain.TransactionStatus `json:"status"`
	At           time.Time                `json:"at"`
	CancelReason string                   `json:"cancel_reason,omitempty"`
	CancelledBy  *uuid.UUID               `json:"cancelled_by,omitempty"`
}

type TransactionCancelDTO struct {
	Reason string `json:"reason"`
}

type TransactionUpdateStatusDTO struct {
	ID     *uuid.UUID               `json:"id"`
	Status domain.TransactionStatus `json:"status"`
}
