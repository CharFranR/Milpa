package ws

import (
	"context"
	"log"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096
)

type Client struct {
	ConversationID uuid.UUID `json:"id"`

	Conn *websocket.Conn

	Message chan *dto.MessageDTO

	MessageUseCase primary.MessageUserCase
}

func (c *Client) WriteMessage() {

	defer func() {
		c.Conn.Close()
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {

		select {

		case message, ok := <-c.Message:

			if !ok {
				return
			}

			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:

			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if err := c.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}

		}
	}

}

func (c *Client) ReadMessage(hub *Hub, ctx context.Context) error {

	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}

	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {

		_, m, err := c.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		msg := dto.MessageDTO{
			ConversationID: c.ConversationID,
			Content:        string(m),
		}

		saved, err := c.MessageUseCase.CreateMessage(ctx, msg)

		if err != nil {
			return err
		}

		hub.Broadcast <- saved

	}
	return nil
}
