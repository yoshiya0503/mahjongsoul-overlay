package models

import "time"

type RoundInfo struct {
	Chang    int `json:"chang"`    // 場風 0=東 1=南 2=西
	Ju       int `json:"ju"`       // 親の席
	Ben      int `json:"ben"`      // 本場
	RiichiBa int `json:"riichiBa"` // 供託リーチ棒
}

type ScoreChange struct {
	Seat  int `json:"seat"`
	Delta int `json:"delta"`
}

type RoundResult struct {
	Round        RoundInfo     `json:"round"`
	ScoreChanges []ScoreChange `json:"scoreChanges"`
	FinalScores  []int         `json:"finalScores"`
	Winner       int           `json:"winner"` // -1 = 流局
	Timestamp    time.Time     `json:"timestamp"`
}

func NewRoundResult(round RoundInfo, scores, deltaScores []int, winner int) RoundResult {
	result := RoundResult{
		Round:       round,
		FinalScores: scores,
		Winner:      winner,
		Timestamp:   time.Now(),
	}
	for i, d := range deltaScores {
		result.ScoreChanges = append(result.ScoreChanges, ScoreChange{Seat: i, Delta: d})
	}
	return result
}
