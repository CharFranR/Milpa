package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatuOptions int

const (
	matched TransactionStatuOptions = iota
	TransactionInProgress
	TransactionCompleted
	TransactionCancelled
)

type Transaction struct {
	ID      uuid.UUID
	MatchID uuid.UUID
	Status  TransactionStatuOptions

	Created_at time.Time
	Updated_at time.Time
}

func NewTransaction(id uuid.UUID, matchID uuid.UUID, matchStatus TransactionStatuOptions) *Transaction {
	return &Transaction{
		ID:         id,
		MatchID:    matchID,
		Status:     matchStatus,
		Created_at: time.Now(),
		Updated_at: time.Now(),
	}
}
