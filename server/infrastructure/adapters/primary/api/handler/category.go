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

type CategoryHandler struct {
	uc primary.CategoryUseCase
}

func NewCategoryHandler(uc primary.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.GetAll(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
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

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req dto.UpdateCategoryRequest
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

func (h *CategoryHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req dto.CategoryStatusRequest
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
