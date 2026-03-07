package models

type AuthGameEvent struct {
	Players []AuthGamePlayer `json:"players"`
}

type AuthGamePlayer struct {
	Name      string `json:"name"`
	Character string `json:"character"`
}

type NewRoundEvent struct {
	Chang  int   `json:"chang"`
	Ju     int   `json:"ju"`
	Ben    int   `json:"ben"`
	Scores []int `json:"scores"`
}

type HuleEvent struct {
	Scores      []int `json:"scores"`
	DeltaScores []int `json:"deltaScores"`
	Winner      int   `json:"winner"`
}

type NoTileEvent struct {
	Scores      []int `json:"scores"`
	DeltaScores []int `json:"deltaScores"`
}

type GameEndEvent struct {
	Scores  []int `json:"scores"`
	Ranks   []int `json:"ranks"`
	DeltaPt int   `json:"deltaPt"`
}
