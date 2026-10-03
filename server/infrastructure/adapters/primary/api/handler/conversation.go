package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
)

type ConversationHandler struct {
	uc primary.ConversationUserUseCase
}

func NewConversationHandler(uc primary.ConversationUserUseCase) *ConversationHandler {
	return &ConversationHandler{uc: uc}
}

func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateConversationDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.CreateConversation(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.ListConversations(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *ConversationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	result, err := h.uc.GetConversation(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *ConversationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	if err := h.uc.DeleteConversation(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
