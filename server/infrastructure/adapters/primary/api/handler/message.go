package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
)

type MessageHandler struct {
	uc primary.MessageUserCase
}

func NewMessageHandler(uc primary.MessageUserCase) *MessageHandler {
	return &MessageHandler{uc: uc}
}

func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.MessageDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.CreateMessage(r.Context(), req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, nil)
}

func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	result, err := h.uc.ListMessage(r.Context(), conversationID)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *MessageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	if err := h.uc.DeleteMessage(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
