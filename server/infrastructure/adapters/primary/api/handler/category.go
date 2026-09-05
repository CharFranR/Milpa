package handler

import (
	"net/http"

	"milpa/domain/port/primary"
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
