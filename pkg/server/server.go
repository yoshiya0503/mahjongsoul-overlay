package server

import (
	"io/fs"
	"net/http"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/game"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/handler"
)

type Server struct {
	app *fiber.App
}

func New(publicFS fs.FS) *Server {
	state := game.NewGameState()
	h := handler.New(state)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/hook", fiberws.New(h.HandleHook))
	app.Get("/ws/overlay", fiberws.New(h.HandleOverlay))

	app.Get("/api/state", h.HandleState)
	app.Post("/api/session/clear", h.HandleClearSession)

	app.Use("/", filesystem.New(filesystem.Config{
		Root: http.FS(publicFS),
	}))

	return &Server{app: app}
}

func (s *Server) ListenAndServe(addr, certFile, keyFile string) error {
	if certFile != "" && keyFile != "" {
		return s.app.ListenTLS(addr, certFile, keyFile)
	}
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() {
	s.app.Shutdown()
}
