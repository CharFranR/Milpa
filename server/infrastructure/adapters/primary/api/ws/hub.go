package ws

import (
	"milpa/aplication/dto"

	"github.com/google/uuid"
)

type Hub struct {
	// Registered clients.
	Clients map[uuid.UUID]map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan *dto.MessageDTO

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *dto.MessageDTO),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[uuid.UUID]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:

			clients, ok := h.Clients[client.ConversationID]

			if !ok {
				clients = make(map[*Client]bool)
				h.Clients[client.ConversationID] = clients

			}

			clients[client] = true

		case client := <-h.Unregister:

			clients, ok := h.Clients[client.ConversationID]

			if !ok {
				continue
			}

			if _, ok = clients[client]; ok {
				delete(clients, client)
				close(client.Message)
			}

			if len(clients) == 0 {
				delete(h.Clients, client.ConversationID)
			}

		case message := <-h.Broadcast:

			clients := h.Clients[message.ConversationID]

			for client := range clients {
				select {
				case client.Message <- message:
				default:
					close(client.Message)
					delete(clients, client)
				}
			}

		}
	}
}
