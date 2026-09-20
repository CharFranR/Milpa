package handler

import (
	"log"
	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"net/http"
	"strconv"
)

type SearchHandler struct {
	uc primary.FuzzyUseCase
}

func NewSearchHandler(fuzzy primary.FuzzyUseCase) *SearchHandler {
	return &SearchHandler{
		uc: fuzzy,
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	req := dto.SearchRequest{
		Term:         r.URL.Query().Get("term"),
		Type:         r.URL.Query().Get("type"),
		Department:   r.URL.Query().Get("department"),
		Municipality: r.URL.Query().Get("municipality"),
		FarmerID:     r.URL.Query().Get("farmer_id"),
		SortBy:       dto.SearchSortField(r.URL.Query().Get("sort")),
	}

	if v := r.URL.Query().Get("price_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.PriceMin = &f
		}
	}
	if v := r.URL.Query().Get("price_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.PriceMax = &f
		}
	}
	if v := r.URL.Query().Get("lat"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.Latitude = f
		}
	}
	if v := r.URL.Query().Get("lng"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.Longitude = f
		}
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			req.Page = i
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			req.PageSize = i
		}
	}

	response, err := h.uc.Search(r.Context(), req)
	if err != nil {
		log.Printf("Search error: %v", err)
		respondError(w, http.StatusBadRequest, "Search error")
		return
	}

	respond(w, http.StatusOK, response)
}
