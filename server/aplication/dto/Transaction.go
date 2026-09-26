package dto

import (
	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type TransactionDTO struct {
	ID      *uuid.UUID               `json:"id"`
	MatchID *uuid.UUID               `json:"match_id"`
	Status  domain.TransactionStatus `json:"status"`
}

type TransactionUpdateStatusDTO struct {
	ID     *uuid.UUID               `json:"id"`
	Status domain.TransactionStatus `json:"status"`
}
