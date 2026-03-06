package main

import (
	"embed"
	"flag"
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
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/game"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/handler"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	addr := flag.String("addr", ":8787", "listen address")
	certFile := flag.String("cert", "server.crt", "TLS certificate file")
	keyFile := flag.String("key", "server.key", "TLS key file")
	flag.Parse()

	webContent, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

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
		Root: http.FS(webContent),
	}))

	go func() {
		log.Printf("mahjongsoul-overlay starting on https://localhost%s", *addr)
		if *certFile != "" && *keyFile != "" {
			if err := app.ListenTLS(*addr, *certFile, *keyFile); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := app.Listen(*addr); err != nil {
				log.Fatal(err)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	app.Shutdown()
}
