package handler

import (
	"encoding/json"
	"log"

	fiberws "github.com/gofiber/contrib/websocket"
)

func (h *Handler) HandleHook(c *fiberws.Conn) {
	defer c.Close()

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			log.Printf("hook read: %v", err)
			return
		}

		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("hook unmarshal: %v", err)
			continue
		}
		if h.State.HandleEvent(msg.Type, msg.Data) {
			h.BroadcastState()
		}
	}
}
