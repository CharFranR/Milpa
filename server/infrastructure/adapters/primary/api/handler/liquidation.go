package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"milpa/internal/validate"
)

type LiquidationHandler struct {
	uc primary.LiquidationUseCase
}

func NewLiquidationHandler(uc primary.LiquidationUseCase) *LiquidationHandler {
	return &LiquidationHandler{uc: uc}
}

func (h *LiquidationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLiquidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "product_name", Value: req.ProductName},
		{Field: "quantity", Value: req.Quantity},
		{Field: "unit_of_measure", Value: req.UnitOfMeasure},
		{Field: "total_price", Value: req.TotalPrice},
		{Field: "unit_price", Value: req.UnitPrice},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.CreateLiquidation(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *LiquidationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid liquidation id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *LiquidationHandler) GetBySupplier(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(r.URL.Query().Get("supplier_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supplier_id")
		return
	}

	result, err := h.uc.GetBySupplier(r.Context(), supplierID)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *LiquidationHandler) GetOpen(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.GetOpen(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *LiquidationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid liquidation id")
		return
	}

	var req dto.UpdateLiquidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateLiquidation(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *LiquidationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid liquidation id")
		return
	}

	if err := h.uc.DeleteLiquidation(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
