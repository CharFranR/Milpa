package handler

import (
	"net/http"

	primary "milpa/domain/port/primary"
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