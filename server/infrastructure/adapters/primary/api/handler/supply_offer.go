package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"milpa/internal/auth"
)

type SupplyOfferHandler struct {
	uc primary.SupplyOfferUseCase
}

func NewSupplyOfferHandler(uc primary.SupplyOfferUseCase) *SupplyOfferHandler {
	return &SupplyOfferHandler{uc: uc}
}

func (h *SupplyOfferHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.SupplyOfferDTO
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

func (h *SupplyOfferHandler) ListBySupplier(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	result, err := h.uc.ListBySupplier(r.Context(), principal.UserID)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyOfferHandler) ListByRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "request_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply request id")
		return
	}

	result, err := h.uc.ListByRequest(r.Context(), requestID)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyOfferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply offer id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplyOfferHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply offer id")
		return
	}

	var req dto.SupplyOfferUpdateDTO
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

func (h *SupplyOfferHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supply offer id")
		return
	}

	if err := h.uc.Withdraw(r.Context(), id); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
