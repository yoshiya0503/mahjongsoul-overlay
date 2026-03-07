package models

type GameState struct {
	InGame    bool            `json:"inGame"`
	Players   Players         `json:"players"`
	Round     RoundInfo       `json:"round"`
	History   []RoundResult   `json:"history"`
	RankPoint RankPointInfo   `json:"rankPoint"`
	Session   []SessionResult `json:"session"`
}

func NewGameState() *GameState {
	return &GameState{
		Players: make(Players, 0, 4),
		History: make([]RoundResult, 0),
		Session: make([]SessionResult, 0),
	}
}
