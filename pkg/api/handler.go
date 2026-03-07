package api

import (
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/services"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/store"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/ws"
)

type Handler struct {
	svc   *services.Game
	store *store.FileStore
	ws    *ws.Handler
}

func New(svc *services.Game, store *store.FileStore, ws *ws.Handler) *Handler {
	return &Handler{svc: svc, store: store, ws: ws}
}
