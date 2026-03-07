package api

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleState(c *fiber.Ctx) error {
	data, err := h.svc.JSON()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("Content-Type", "application/json")
	return c.Send(data)
}

func (h *Handler) HandleClearSession(c *fiber.Ctx) error {
	h.svc.ClearSession()
	if err := h.store.SaveSession(h.svc.Session()); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	h.ws.BroadcastState()
	return c.JSON(fiber.Map{"ok": true})
}
