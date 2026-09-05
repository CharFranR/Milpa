package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"milpa/internal/validate"
)

type ReviewHandler struct {
	uc primary.ReviewUseCase
}

func NewReviewHandler(uc primary.ReviewUseCase) *ReviewHandler {
	return &ReviewHandler{uc: uc}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "company_id", Value: req.CompanyID},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.CreateReview(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	if companyStr := r.URL.Query().Get("company_id"); companyStr != "" {
		companyID, err := uuid.Parse(companyStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid company_id")
			return
		}

		result, err := h.uc.FindByCompany(r.Context(), companyID)
		if err != nil {
			handleError(w, err)
			return
		}

		respond(w, http.StatusOK, result)
		return
	}

	if userStr := r.URL.Query().Get("user_id"); userStr != "" {
		userID, err := uuid.Parse(userStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		result, err := h.uc.FindByUser(r.Context(), userID)
		if err != nil {
			handleError(w, err)
			return
		}

		respond(w, http.StatusOK, result)
		return
	}

	respondError(w, http.StatusBadRequest, "provide company_id or user_id")
}
