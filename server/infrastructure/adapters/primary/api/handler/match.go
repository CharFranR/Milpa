package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
)

type MatchHandler struct {
	uc primary.MatchUseCase
}

func NewMatchHandler(uc primary.MatchUseCase) *MatchHandler {
	return &MatchHandler{uc: uc}
}

func (h *MatchHandler) Like(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "offerID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offer id")
		return
	}

	match, transaction, err := h.uc.Like(r.Context(), offerID)
	if err != nil {
		handleMatchError(w, err)
		return
	}

	respond(w, http.StatusCreated, dto.MatchCreatedDTO{Match: match, Transaction: transaction})
}

func (h *MatchHandler) Pass(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "offerID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offer id")
		return
	}

	if err := h.uc.Pass(r.Context(), offerID); err != nil {
		handleMatchError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *MatchHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	matchID, err := uuid.Parse(chi.URLParam(r, "matchID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), matchID)
	if err != nil {
		handleMatchError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *MatchHandler) ListByRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "requestID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	result, err := h.uc.ListByRequest(r.Context(), requestID)
	if err != nil {
		handleMatchError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *MatchHandler) ListPrioritized(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "requestID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	result, err := h.uc.ListPrioritized(r.Context(), requestID)
	if err != nil {
		handleMatchError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func handleMatchError(w http.ResponseWriter, err error) {
	if isMatchConflict(err) {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	handleError(w, err)
}

func isMatchConflict(err error) bool {
	return errors.Is(err, domain.ErrInvalidOfferStatus) ||
		errors.Is(err, domain.ErrInvalidRequestStatus) ||
		errors.Is(err, domain.ErrInvalidMatchStatus) ||
		errors.Is(err, domain.ErrInsufficientAmount) ||
		errors.Is(err, domain.ErrInvalidTransactionTransition)
}
