package ws

import (
	"log"
	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api/httpx"
	"milpa/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub             *Hub
	uc_message      primary.MessageUserCase
	uc_conversation primary.ConversationUserUseCase
}

func NewHandler(h *Hub, uc_message primary.MessageUserCase, uc_conversation primary.ConversationUserUseCase) *Handler {
	return &Handler{
		hub:             h,
		uc_message:      uc_message,
		uc_conversation: uc_conversation,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{auth.WebSocketProtocol},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		httpx.RespondError(w, status, reason.Error())
	},
}

func (h *Handler) WSHandler(w http.ResponseWriter, r *http.Request) {

	_, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "conversationID"))

	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.uc_conversation.GetConversation(r.Context(), conversationID); err != nil {
		httpx.RespondError(w, httpx.StatusCode(err), err.Error())
		return

	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	cl1 := &Client{
		ConversationID: conversationID,
		Conn:           conn,
		Message:        make(chan *dto.MessageDTO, 10),
		MessageUseCase: h.uc_message,
	}

	msg := &dto.MessageDTO{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Content:        "Espacio disponible gracias a Hackaton Nicaragua",
	}

	h.hub.Register <- cl1

	// Mensaje sponsor
	h.hub.Broadcast <- msg

	go cl1.WriteMessage()

	if err := cl1.ReadMessage(h.hub, r.Context()); err != nil {
		log.Printf("ws: read loop ended: %v", err)
	}
}
