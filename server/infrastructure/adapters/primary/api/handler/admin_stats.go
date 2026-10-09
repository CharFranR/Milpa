package handler

import (
	"net/http"

	primary "milpa/domain/port/primary"
)

type AdminStatsHandler struct {
	uc primary.AdminStatsUseCase
}

func NewAdminStatsHandler(uc primary.AdminStatsUseCase) *AdminStatsHandler {
	return &AdminStatsHandler{uc: uc}
}

func (h *AdminStatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.Get(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}
