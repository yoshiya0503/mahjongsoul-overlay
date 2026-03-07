package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/api"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/config"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/services"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/store"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/ws"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	config.Init()

	addr := config.Addr()

	webContent, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	st := store.NewFileStore(config.SessionFile())
	session, _ := st.LoadSession()
	svc := services.NewGame(session)
	wsHandler := ws.New(svc, st)
	apiHandler := api.New(svc, st, wsHandler)

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

	app.Get("/ws/hook", fiberws.New(wsHandler.HandleHook))
	app.Get("/ws/overlay", fiberws.New(wsHandler.HandleOverlay))

	app.Get("/api/state", apiHandler.HandleState)
	app.Post("/api/session/clear", apiHandler.HandleClearSession)

	app.Use("/", filesystem.New(filesystem.Config{
		Root: http.FS(webContent),
	}))

	go func() {
		log.Printf("mahjongsoul-overlay starting on http://localhost%s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	app.Shutdown()
}
