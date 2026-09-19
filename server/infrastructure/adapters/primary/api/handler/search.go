package handler

import (
	"log"
	"milpa/domain/port/primary"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SearchHandler struct {
	uc primary.FuzzyUseCase
}

func NewSearchHandler(fuzzy primary.FuzzyUseCase) *SearchHandler {
	return &SearchHandler{
		uc: fuzzy,
	}
}

func (h *SearchHandler) List(w http.ResponseWriter, r *http.Request) {

	term := chi.URLParam(r, "term")

	response, err := h.uc.Search(r.Context(), term)

	if err != nil {
		log.Println("List error: %w", err)
		respondError(w, http.StatusBadRequest, "List error")
		return
	}

	respond(w, http.StatusOK, response)
}
