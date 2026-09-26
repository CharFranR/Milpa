package handler

import (
	"net/http"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
)

type RecommendationHandler struct {
	uc primary.RecommendationUseCase
}

func NewRecommendationHandler(uc primary.RecommendationUseCase) *RecommendationHandler {
	return &RecommendationHandler{uc: uc}
}

func (h *RecommendationHandler) Availability(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(r.URL.Query().Get("supplier_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supplier id")
		return
	}

	productName := r.URL.Query().Get("product_name")
	if productName == "" {
		respondError(w, http.StatusBadRequest, "product_name is required")
		return
	}

	available, err := h.uc.AvailableQuantity(r.Context(), supplierID, productName)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, dto.AvailabilityDTO{
		SupplierID:        supplierID,
		ProductName:       productName,
		AvailableQuantity: available,
	})
}
