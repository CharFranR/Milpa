package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	primary "milpa/domain/port/primary"
	"milpa/internal/validate"
)

type UnitOfMeasureHandler struct {
	uc primary.UnitOfMeasureUseCase
}

func NewUnitOfMeasureHandler(uc primary.UnitOfMeasureUseCase) *UnitOfMeasureHandler {
	return &UnitOfMeasureHandler{uc: uc}
}

func (h *UnitOfMeasureHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *UnitOfMeasureHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.ListAll(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *UnitOfMeasureHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUnitOfMeasureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "code", Value: req.Code},
		{Field: "name", Value: req.Name},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.Create(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *UnitOfMeasureHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid unit of measure id")
		return
	}

	var req dto.UpdateUnitOfMeasureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.Update(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *UnitOfMeasureHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid unit of measure id")
		return
	}

	var req dto.UnitOfMeasureStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.SetStatus(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}
