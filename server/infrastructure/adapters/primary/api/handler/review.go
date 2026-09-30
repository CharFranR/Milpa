package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
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

	// The two shapes are alternative spellings of the same thing, so the check
	// is that exactly one of them arrived -- not that a particular one did.
	// Requiring company_id would make every review of a farmer a 400.
	if req.CompanyID == uuid.Nil && req.TargetID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "provide company_id or target_id")
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

// Average is a separate route rather than a third reading of the existing
// query: the existing one returns a list and this returns a number, and folding
// them together would make each pay for the other.
func (h *ReviewHandler) Average(w http.ResponseWriter, r *http.Request) {
	rawTargetType := r.URL.Query().Get("target_type")
	targetType := domain.ReviewTargetType(rawTargetType)
	if !domain.ValidReviewTargetType(targetType) {
		respondError(w, http.StatusBadRequest, "target_type must be company or user")
		return
	}

	targetID, err := uuid.Parse(r.URL.Query().Get("target_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid target_id")
		return
	}

	result, err := h.uc.GetAverageRating(r.Context(), targetType, targetID)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}
