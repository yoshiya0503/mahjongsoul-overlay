package handler

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleState(c *fiber.Ctx) error {
	data, err := h.State.JSON()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("Content-Type", "application/json")
	return c.Send(data)
}

func (h *Handler) HandleClearSession(c *fiber.Ctx) error {
	h.State.ClearSession()
	h.BroadcastState()
	return c.JSON(fiber.Map{"ok": true})
}
