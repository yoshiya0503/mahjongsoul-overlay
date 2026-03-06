package handler

import (
	"log"

	"github.com/fasthttp/websocket"
	fiberws "github.com/gofiber/contrib/websocket"
)

func (h *Handler) HandleOverlay(c *fiberws.Conn) {
	conn := c.Conn
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		c.Close()
	}()

	if data, err := h.State.JSON(); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}

	log.Println("overlay connected")
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Handler) BroadcastState() {
	data, err := h.State.JSON()
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("broadcast: %v", err)
		}
	}
}
