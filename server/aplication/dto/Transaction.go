package dto

import (
	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type TransactionDTO struct {
	ID          *uuid.UUID                     `json:"id"`
	MatchID     *uuid.UUID                     `json:"match_id"`
	MatchStatus domain.TransactionStatuOptions `json:"status"`
}

type TransactionUpdateStatusDTO struct {
	ID          *uuid.UUID                     `json:"id"`
	MatchStatus domain.TransactionStatuOptions `json:"status"`
}
