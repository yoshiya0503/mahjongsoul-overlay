package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/fasthttp/websocket"
	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/yoshiya0503/jantama-overlay/services/game"
)

type Server struct {
	app     *fiber.App
	state   *game.GameState
	clients map[*websocket.Conn]struct{}
	mu      sync.RWMutex
}

func New(webFS fs.FS) *Server {
	s := &Server{
		state:   game.NewGameState(),
		clients: make(map[*websocket.Conn]struct{}),
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/hook", fiberws.New(s.handleHook))
	app.Get("/ws/overlay", fiberws.New(s.handleOverlay))

	app.Get("/api/state", s.handleAPIState)
	app.Post("/api/session/clear", s.handleClearSession)

	app.Use("/", filesystem.New(filesystem.Config{
		Root: http.FS(webFS),
	}))

	s.app = app
	return s
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

func (s *Server) handleHook(c *fiberws.Conn) {
	log.Println("userscript connected")
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
		log.Printf("event: %s", msg.Type)

		if s.state.HandleEvent(msg.Type, msg.Data) {
			s.broadcastState()
		}
	}
}

func (s *Server) handleOverlay(c *fiberws.Conn) {
	conn := c.Conn
	s.mu.Lock()
	s.clients[conn] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		c.Close()
	}()

	if data, err := s.state.JSON(); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}

	log.Println("overlay connected")
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *Server) handleAPIState(c *fiber.Ctx) error {
	data, err := s.state.JSON()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("Content-Type", "application/json")
	return c.Send(data)
}

func (s *Server) handleClearSession(c *fiber.Ctx) error {
	s.state.ClearSession()
	s.broadcastState()
	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) broadcastState() {
	data, err := s.state.JSON()
	if err != nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("broadcast: %v", err)
		}
	}
}
