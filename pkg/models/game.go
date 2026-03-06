package models

import "time"

type Player struct {
	Seat      int    `json:"seat"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Rank      int    `json:"rank"`
	Character string `json:"character"`
}

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

type RankPointInfo struct {
	CurrentPt int    `json:"currentPt"`
	TargetPt  int    `json:"targetPt"`
	RankName  string `json:"rankName"`
}

type SessionResult struct {
	Rank      int       `json:"rank"`
	Score     int       `json:"score"`
	DeltaPt   int       `json:"deltaPt"`
	Timestamp time.Time `json:"timestamp"`
}
