package handler

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/game"
)

type Handler struct {
	State   *game.GameState
	clients map[*websocket.Conn]struct{}
	mu      sync.RWMutex
}

func New(state *game.GameState) *Handler {
	return &Handler{
		State:   state,
		clients: make(map[*websocket.Conn]struct{}),
	}
}
