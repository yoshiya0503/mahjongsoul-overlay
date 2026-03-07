package ws

import (
	"encoding/json"
	"log"

	"github.com/fasthttp/websocket"
	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/services"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/store"
)

type hookMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Handler struct {
	svc   *services.Game
	store *store.FileStore
	hub   *Hub
}

func New(svc *services.Game, store *store.FileStore) *Handler {
	return &Handler{
		svc:   svc,
		store: store,
		hub:   NewHub(),
	}
}

func (h *Handler) HandleHook(c *fiberws.Conn) {
	defer c.Close()

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			log.Printf("hook read: %v", err)
			return
		}

		var msg hookMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("hook unmarshal: %v", err)
			continue
		}
		if h.svc.HandleEvent(msg.Type, msg.Data) {
			h.saveSession()
			h.BroadcastState()
		}
	}
}

func (h *Handler) HandleOverlay(c *fiberws.Conn) {
	conn := c.Conn
	h.hub.Add(conn)
	defer func() {
		h.hub.Remove(conn)
		c.Close()
	}()

	if data, err := h.svc.JSON(); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Handler) BroadcastState() {
	data, err := h.svc.JSON()
	if err != nil {
		return
	}
	h.hub.Broadcast(data)
}

func (h *Handler) saveSession() {
	if err := h.store.SaveSession(h.svc.Session()); err != nil {
		log.Printf("session save error: %v", err)
	}
}
