package server

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/yoshiya0503/jantama-overlay/services/game"
)

type Server struct {
	state   *game.GameState
	clients map[*websocket.Conn]struct{}
	mu      sync.RWMutex
	httpSrv *http.Server
	webFS   fs.FS
}

func New(webFS fs.FS) *Server {
	return &Server{
		state:   game.NewGameState(),
		clients: make(map[*websocket.Conn]struct{}),
		webFS:   webFS,
	}
}

func (s *Server) ListenAndServe(addr, certFile, keyFile string) error {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(s.webFS)))
	mux.HandleFunc("/ws/hook", s.handleHook)
	mux.HandleFunc("/ws/overlay", s.handleOverlay)
	mux.HandleFunc("/api/state", s.handleAPIState)

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: withCORS(mux),
	}

	if certFile != "" && keyFile != "" {
		return s.httpSrv.ListenAndServeTLS(certFile, keyFile)
	}
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleHook(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("hook accept: %v", err)
		return
	}
	defer conn.CloseNow()
	log.Println("userscript connected")

	for {
		_, data, err := conn.Read(r.Context())
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

func (s *Server) handleOverlay(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("overlay accept: %v", err)
		return
	}
	defer conn.CloseNow()

	s.mu.Lock()
	s.clients[conn] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	if data, err := s.state.JSON(); err == nil {
		conn.Write(r.Context(), websocket.MessageText, data)
	}

	log.Println("overlay connected")
	for {
		if _, _, err := conn.Read(r.Context()); err != nil {
			return
		}
	}
}

func (s *Server) handleAPIState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := s.state.JSON()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}

func (s *Server) broadcastState() {
	data, err := s.state.JSON()
	if err != nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for conn := range s.clients {
		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			log.Printf("broadcast: %v", err)
		}
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
