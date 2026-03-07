package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/config"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

type eventHandler func(json.RawMessage) bool

type Game struct {
	mu       sync.RWMutex
	state    *models.GameState
	handlers map[string]eventHandler
}

func NewGame(session []models.SessionResult) *Game {
	state := models.NewGameState()
	state.Session = session
	g := &Game{state: state}
	g.handlers = map[string]eventHandler{
		"authGame":  g.handleAuthGame,
		"newRound":  g.handleNewRound,
		"hule":      g.handleHule,
		"noTile":    g.handleNoTile,
		"liuju":     g.handleLiuju,
		"gameEnd":   g.handleGameEnd,
		"rankPoint": g.handleRankPoint,
	}
	return g
}

func (g *Game) ClearSession() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.state.Session = g.state.Session[:0]
}

func (g *Game) Session() []models.SessionResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cp := make([]models.SessionResult, len(g.state.Session))
	copy(cp, g.state.Session)
	return cp
}

func (g *Game) JSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return json.Marshal(g.state)
}

func (g *Game) HandleEvent(eventType string, data json.RawMessage) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if h, ok := g.handlers[eventType]; ok {
		return h(data)
	}
	return false
}

func (g *Game) handleAuthGame(data json.RawMessage) bool {
	var ev models.AuthGameEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	g.state.InGame = true
	g.state.Players = make(models.Players, len(ev.Players))
	g.state.History = g.state.History[:0]
	for i, p := range ev.Players {
		g.state.Players[i] = models.Player{
			Seat:      i,
			Name:      p.Name,
			Score:     config.InitialScore(),
			Rank:      i + 1,
			Character: p.Character,
		}
	}
	return true
}

func (g *Game) handleNewRound(data json.RawMessage) bool {
	var ev models.NewRoundEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	g.state.Round = models.RoundInfo{Chang: ev.Chang, Ju: ev.Ju, Ben: ev.Ben}
	g.state.Players.UpdateScores(ev.Scores)
	return true
}

func (g *Game) handleHule(data json.RawMessage) bool {
	var ev models.HuleEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	return g.recordRoundResult(ev.Scores, ev.DeltaScores, ev.Winner)
}

func (g *Game) handleNoTile(data json.RawMessage) bool {
	var ev models.NoTileEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	return g.recordRoundResult(ev.Scores, ev.DeltaScores, -1)
}

func (g *Game) recordRoundResult(scores, deltaScores []int, winner int) bool {
	g.state.History = append(g.state.History, models.NewRoundResult(g.state.Round, scores, deltaScores, winner))
	g.state.Players.UpdateScores(scores)
	return true
}

func (g *Game) handleLiuju(_ json.RawMessage) bool {
	g.state.History = append(g.state.History, models.RoundResult{
		Round:     g.state.Round,
		Winner:    -1,
		Timestamp: time.Now(),
	})
	return true
}

func (g *Game) handleGameEnd(data json.RawMessage) bool {
	if !g.state.InGame {
		return false
	}
	var ev models.GameEndEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	g.state.Players.UpdateScores(ev.Scores)
	myRank := 1
	if len(ev.Ranks) > 0 {
		myRank = ev.Ranks[0]
	}
	g.state.Session = append(g.state.Session, models.SessionResult{
		Rank:      myRank,
		Score:     g.state.Players[0].Score,
		DeltaPt:   ev.DeltaPt,
		Timestamp: time.Now(),
	})
	g.state.InGame = false
	return true
}

func (g *Game) handleRankPoint(data json.RawMessage) bool {
	var rp models.RankPointInfo
	if err := json.Unmarshal(data, &rp); err != nil {
		return false
	}
	g.state.RankPoint = rp
	return true
}
