package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus int

const (
	TransactionMatched TransactionStatus = iota
	TransactionInProgress
	TransactionCompleted
	TransactionCancelled
)

type TransactionParticipant int

const (
	BuyerParticipant TransactionParticipant = iota
	SupplierParticipant
)

type Transaction struct {
	ID                          uuid.UUID
	MatchID                     uuid.UUID
	Status                      TransactionStatus
	BuyerStartConfirmedAt       *time.Time
	SupplierStartConfirmedAt    *time.Time
	BuyerDeliveryConfirmedAt    *time.Time
	SupplierDeliveryConfirmedAt *time.Time
	CancelledBy                 *uuid.UUID
	CancelReason                string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTransaction(matchID uuid.UUID) *Transaction {
	return &Transaction{
		ID:        uuid.New(),
		MatchID:   matchID,
		Status:    TransactionMatched,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t Transaction) IsActive() bool {
	return t.Status == TransactionMatched || t.Status == TransactionInProgress
}

func (t *Transaction) ConfirmStart(p TransactionParticipant, now time.Time) error {
	if t.Status != TransactionMatched {
		return ErrInvalidTransactionTransition
	}
	if t.startConfirmedAt(p) != nil {
		return ErrAlreadyConfirmed
	}
	t.setStartConfirmed(p, now)
	if t.startConfirmedAt(BuyerParticipant) != nil && t.startConfirmedAt(SupplierParticipant) != nil {
		t.Status = TransactionInProgress
	}
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) ConfirmDelivery(p TransactionParticipant, now time.Time) error {
	if t.Status != TransactionInProgress {
		return ErrInvalidTransactionTransition
	}
	if t.deliveryConfirmedAt(p) != nil {
		return ErrAlreadyConfirmed
	}
	t.setDeliveryConfirmed(p, now)
	if t.deliveryConfirmedAt(BuyerParticipant) != nil && t.deliveryConfirmedAt(SupplierParticipant) != nil {
		t.Status = TransactionCompleted
	}
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) Cancel(cancelledBy uuid.UUID, reason string, now time.Time) error {
	if t.Status == TransactionCompleted || t.Status == TransactionCancelled {
		return ErrTerminalState
	}
	if reason == "" {
		return ErrReasonRequired
	}
	t.Status = TransactionCancelled
	t.CancelledBy = &cancelledBy
	t.CancelReason = reason
	t.UpdatedAt = now
	return nil
}

func (t Transaction) startConfirmedAt(p TransactionParticipant) *time.Time {
	if p == BuyerParticipant {
		return t.BuyerStartConfirmedAt
	}
	return t.SupplierStartConfirmedAt
}

func (t *Transaction) setStartConfirmed(p TransactionParticipant, now time.Time) {
	if p == BuyerParticipant {
		t.BuyerStartConfirmedAt = &now
		return
	}
	t.SupplierStartConfirmedAt = &now
}

func (t Transaction) deliveryConfirmedAt(p TransactionParticipant) *time.Time {
	if p == BuyerParticipant {
		return t.BuyerDeliveryConfirmedAt
	}
	return t.SupplierDeliveryConfirmedAt
}

func (t *Transaction) setDeliveryConfirmed(p TransactionParticipant, now time.Time) {
	if p == BuyerParticipant {
		t.BuyerDeliveryConfirmedAt = &now
		return
	}
	t.SupplierDeliveryConfirmedAt = &now
}

func (t TransactionStatus) String() string {
	switch t {
	case TransactionMatched:
		return "matched"
	case TransactionInProgress:
		return "in_progress"
	case TransactionCompleted:
		return "completed"
	case TransactionCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}
