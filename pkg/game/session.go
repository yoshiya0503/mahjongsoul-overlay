package game

import (
	"encoding/json"
	"log"
	"os"

	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/config"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

func (gs *GameState) saveSession() {
	data, err := json.Marshal(gs.Session)
	if err != nil {
		log.Printf("session save error: %v", err)
		return
	}
	if err := os.WriteFile(config.SessionFile(), data, 0644); err != nil {
		log.Printf("session save error: %v", err)
	}
}

func (gs *GameState) loadSession() {
	data, err := os.ReadFile(config.SessionFile())
	if err != nil {
		return
	}
	var session []models.SessionResult
	if err := json.Unmarshal(data, &session); err != nil {
		log.Printf("session load error: %v", err)
		return
	}
	gs.Session = session
	log.Printf("session loaded: %d games", len(session))
}
