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

type SupplierInventoryHandler struct {
	uc primary.SupplierInventoryUseCase
}

func NewSupplierInventoryHandler(uc primary.SupplierInventoryUseCase) *SupplierInventoryHandler {
	return &SupplierInventoryHandler{uc: uc}
}

func (h *SupplierInventoryHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	var req dto.UpsertSupplierInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.Upsert(r.Context(), principal.UserID, req)
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *SupplierInventoryHandler) ListBySupplier(w http.ResponseWriter, r *http.Request) {
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

func (h *SupplierInventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		handleSupplyError(w, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	if err := h.uc.Delete(r.Context(), principal.UserID, id); err != nil {
		handleSupplyError(w, err)
		return
	}

	respond(w, http.StatusNoContent, nil)
}
