package game

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

type GameState struct {
	mu sync.RWMutex

	InGame    bool                   `json:"inGame"`
	Players   []models.Player        `json:"players"`
	Round     models.RoundInfo       `json:"round"`
	History   []models.RoundResult   `json:"history"`
	RankPoint models.RankPointInfo   `json:"rankPoint"`
	Session   []models.SessionResult `json:"session"`
}

func NewGameState() *GameState {
	gs := &GameState{
		Players: make([]models.Player, 0, 4),
		History: make([]models.RoundResult, 0),
		Session: make([]models.SessionResult, 0),
	}
	gs.loadSession()
	return gs
}

func (gs *GameState) ClearSession() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Session = gs.Session[:0]
	gs.saveSession()
}

func (gs *GameState) JSON() ([]byte, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return json.Marshal(gs)
}

func (gs *GameState) HandleEvent(eventType string, data json.RawMessage) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	switch eventType {
	case "authGame":
		return gs.handleAuthGame(data)
	case "newRound":
		return gs.handleNewRound(data)
	case "hule":
		return gs.handleHule(data)
	case "noTile":
		return gs.handleNoTile(data)
	case "liuju":
		return gs.handleLiuju(data)
	case "gameEnd":
		return gs.handleGameEnd(data)
	case "rankPoint":
		return gs.handleRankPoint(data)
	default:
		return false
	}
}

type authGameEvent struct {
	Players []struct {
		Name      string `json:"name"`
		Character string `json:"character"`
	} `json:"players"`
}

func (gs *GameState) handleAuthGame(data json.RawMessage) bool {
	var ev authGameEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	gs.InGame = true
	gs.Players = make([]models.Player, len(ev.Players))
	gs.History = gs.History[:0]
	for i, p := range ev.Players {
		gs.Players[i] = models.Player{
			Seat:      i,
			Name:      p.Name,
			Score:     25000,
			Rank:      i + 1,
			Character: p.Character,
		}
	}
	return true
}

type newRoundEvent struct {
	Chang  int   `json:"chang"`
	Ju     int   `json:"ju"`
	Ben    int   `json:"ben"`
	Scores []int `json:"scores"`
}

func (gs *GameState) handleNewRound(data json.RawMessage) bool {
	var ev newRoundEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	gs.Round = models.RoundInfo{Chang: ev.Chang, Ju: ev.Ju, Ben: ev.Ben}
	if len(ev.Scores) == len(gs.Players) {
		for i := range gs.Players {
			gs.Players[i].Score = ev.Scores[i]
		}
		gs.updateRanks()
	}
	return true
}

type huleEvent struct {
	Scores      []int `json:"scores"`
	DeltaScores []int `json:"deltaScores"`
	Winner      int   `json:"winner"`
}

func (gs *GameState) handleHule(data json.RawMessage) bool {
	var ev huleEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	result := models.RoundResult{
		Round:       gs.Round,
		FinalScores: ev.Scores,
		Winner:      ev.Winner,
		Timestamp:   time.Now(),
	}
	for i, d := range ev.DeltaScores {
		result.ScoreChanges = append(result.ScoreChanges, models.ScoreChange{Seat: i, Delta: d})
	}
	gs.History = append(gs.History, result)
	if len(ev.Scores) == len(gs.Players) {
		for i := range gs.Players {
			gs.Players[i].Score = ev.Scores[i]
		}
		gs.updateRanks()
	}
	return true
}

type noTileEvent struct {
	Scores      []int `json:"scores"`
	DeltaScores []int `json:"deltaScores"`
}

func (gs *GameState) handleNoTile(data json.RawMessage) bool {
	var ev noTileEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	result := models.RoundResult{
		Round:       gs.Round,
		FinalScores: ev.Scores,
		Winner:      -1,
		Timestamp:   time.Now(),
	}
	for i, d := range ev.DeltaScores {
		result.ScoreChanges = append(result.ScoreChanges, models.ScoreChange{Seat: i, Delta: d})
	}
	gs.History = append(gs.History, result)
	if len(ev.Scores) == len(gs.Players) {
		for i := range gs.Players {
			gs.Players[i].Score = ev.Scores[i]
		}
		gs.updateRanks()
	}
	return true
}

func (gs *GameState) handleLiuju(_ json.RawMessage) bool {
	gs.History = append(gs.History, models.RoundResult{
		Round:     gs.Round,
		Winner:    -1,
		Timestamp: time.Now(),
	})
	return true
}

type gameEndEvent struct {
	Scores  []int `json:"scores"`
	Ranks   []int `json:"ranks"`
	DeltaPt int   `json:"deltaPt"`
}

func (gs *GameState) handleGameEnd(data json.RawMessage) bool {
	if !gs.InGame {
		return false
	}
	var ev gameEndEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return false
	}
	if len(ev.Scores) == len(gs.Players) {
		for i := range gs.Players {
			gs.Players[i].Score = ev.Scores[i]
		}
		gs.updateRanks()
	}
	myRank := 1
	if len(ev.Ranks) > 0 {
		myRank = ev.Ranks[0]
	}
	gs.Session = append(gs.Session, models.SessionResult{
		Rank:      myRank,
		Score:     gs.Players[0].Score,
		DeltaPt:   ev.DeltaPt,
		Timestamp: time.Now(),
	})
	gs.InGame = false
	gs.saveSession()
	return true
}

func (gs *GameState) handleRankPoint(data json.RawMessage) bool {
	var rp models.RankPointInfo
	if err := json.Unmarshal(data, &rp); err != nil {
		return false
	}
	gs.RankPoint = rp
	return true
}

func (gs *GameState) updateRanks() {
	for i := range gs.Players {
		rank := 1
		for j := range gs.Players {
			if i != j && (gs.Players[j].Score > gs.Players[i].Score ||
				(gs.Players[j].Score == gs.Players[i].Score && gs.Players[j].Seat < gs.Players[i].Seat)) {
				rank++
			}
		}
		gs.Players[i].Rank = rank
	}
}
