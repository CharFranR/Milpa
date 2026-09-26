package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
)

type TransactionHandler struct {
	uc primary.TransactionUseCase
}

func NewTransactionHandler(uc primary.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) GetByMatch(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "match_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	result, err := h.uc.GetByMatch(r.Context(), matchID)
	if err != nil {
		handleTransactionError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *TransactionHandler) ListByRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "request_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	result, err := h.uc.ListByRequest(r.Context(), requestID)
	if err != nil {
		handleTransactionError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *TransactionHandler) ConfirmStart(w http.ResponseWriter, r *http.Request) {
	transactionID, err := uuid.Parse(chi.URLParam(r, "transaction_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := h.uc.ConfirmStart(r.Context(), transactionID); err != nil {
		handleTransactionError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *TransactionHandler) ConfirmDelivery(w http.ResponseWriter, r *http.Request) {
	transactionID, err := uuid.Parse(chi.URLParam(r, "transaction_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := h.uc.ConfirmDelivery(r.Context(), transactionID); err != nil {
		handleTransactionError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *TransactionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	transactionID, err := uuid.Parse(chi.URLParam(r, "transaction_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var req dto.TransactionCancelDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.Cancel(r.Context(), transactionID, req.Reason); err != nil {
		handleTransactionError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func handleTransactionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTerminalState),
		errors.Is(err, domain.ErrInvalidTransactionTransition),
		errors.Is(err, domain.ErrAlreadyConfirmed),
		errors.Is(err, domain.ErrInvalidMatchStatus),
		errors.Is(err, domain.ErrInvalidRequestStatus),
		errors.Is(err, domain.ErrInvalidOfferStatus):
		respondError(w, http.StatusConflict, err.Error())
	default:
		handleError(w, err)
	}
}
