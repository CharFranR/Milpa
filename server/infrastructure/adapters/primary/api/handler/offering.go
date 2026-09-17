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

type OfferingHandler struct {
	uc primary.OfferingUseCase
}

func NewOfferingHandler(uc primary.OfferingUseCase) *OfferingHandler {
	return &OfferingHandler{uc: uc}
}

func (h *OfferingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "user_id", Value: req.UserID},
		{Field: "name", Value: req.Name},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.CreateOffering(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *OfferingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offering id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *OfferingHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	result, err := h.uc.GetByUserID(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *OfferingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offering id")
		return
	}

	var req dto.UpdateOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateOffering(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
