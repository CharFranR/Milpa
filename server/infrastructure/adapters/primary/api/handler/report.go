package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
)

type ReportHandler struct {
	uc primary.ReportUseCase
}

func NewReportHandler(uc primary.ReportUseCase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.Create(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	targetType := r.URL.Query().Get("target_type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := h.uc.List(r.Context(), status, targetType, page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *ReportHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid report id")
		return
	}

	var req dto.ResolveReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.Resolve(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

type ModerationHandler struct {
	uc primary.ModerationUseCase
}

func NewModerationHandler(uc primary.ModerationUseCase) *ModerationHandler {
	return &ModerationHandler{uc: uc}
}

func (h *ModerationHandler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dto.SuspendUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.SuspendUser(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *ModerationHandler) DeleteOffering(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offering id")
		return
	}

	if err := h.uc.DeleteOffering(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusNoContent, nil)
}

func (h *ModerationHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	actorID := r.URL.Query().Get("actor_id")
	targetType := r.URL.Query().Get("target_type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := h.uc.ListAuditLogs(r.Context(), action, actorID, targetType, page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}
