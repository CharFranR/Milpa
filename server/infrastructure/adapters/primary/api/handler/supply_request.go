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
	"milpa/internal/auth"
)

type SupplyRequestHandler struct {
	uc primary.SupplyRequestUseCase
}

func NewSupplyRequestHandler(uc primary.SupplyRequestUseCase) *SupplyRequestHandler {
	return &SupplyRequestHandler{uc: uc}
}

func supplyErrorIsConflict(err error) bool {
	return errors.Is(err, domain.ErrInvalidRequestStatus) ||
		errors.Is(err, domain.ErrInvalidOfferStatus) ||
		errors.Is(err, domain.ErrInsufficientAmount) ||
		errors.Is(err, domain.ErrDuplicate) ||
		errors.Is(err, primary.ErrActiveMatch)
}

func handleSupplyError(w http.ResponseWriter, err error) {
	if supplyErrorIsConflict(err) {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	handleError(w, err)
}

func (h *SupplyRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.SupplyRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.Create(r.Context(), req)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *SupplyRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.List(r.Context())
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyRequestHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	result, err := h.uc.ListAvailable(r.Context(), principal.UserID)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyRequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyRequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	var req dto.SupplyGeneralUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.Update(r.Context(), id, req); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *SupplyRequestHandler) UpdateAmounts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	var req dto.SupplyUpdateAmountsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateAmounts(r.Context(), id, req); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *SupplyRequestHandler) UpdateDeadlines(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	var req dto.SupplyUpdateTimeDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateDeadlines(r.Context(), id, req); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *SupplyRequestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	if err := h.uc.Cancel(r.Context(), id); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *SupplyRequestHandler) Expire(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	if err := h.uc.Expire(r.Context(), id); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
