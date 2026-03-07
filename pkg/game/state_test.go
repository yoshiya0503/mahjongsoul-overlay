package game

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/config"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

func TestMain(m *testing.M) {
	config.Init()
	os.Exit(m.Run())
}

func newTestState() *GameState {
	return &GameState{
		Players: make(models.Players, 0, 4),
		History: make([]models.RoundResult, 0),
		Session: make([]models.SessionResult, 0),
	}
}

func TestHandleEvent_UnknownType(t *testing.T) {
	gs := newTestState()
	assert.False(t, gs.HandleEvent("unknown", json.RawMessage(`{}`)))
}

func TestHandleAuthGame(t *testing.T) {
	gs := newTestState()
	data := json.RawMessage(`{"players":[{"name":"Alice","character":"1"},{"name":"Bob","character":"2"},{"name":"Carol","character":"3"},{"name":"Dave","character":"4"}]}`)

	assert.True(t, gs.HandleEvent("authGame", data))
	assert.True(t, gs.InGame)
	assert.Len(t, gs.Players, 4)
	assert.Equal(t, "Alice", gs.Players[0].Name)
	assert.Equal(t, "Dave", gs.Players[3].Name)
	assert.Equal(t, 25000, gs.Players[0].Score)
	assert.Equal(t, 0, gs.Players[0].Seat)
	assert.Equal(t, 1, gs.Players[0].Rank)
}

func TestHandleAuthGame_InvalidJSON(t *testing.T) {
	gs := newTestState()
	assert.False(t, gs.HandleEvent("authGame", json.RawMessage(`invalid`)))
}

func TestHandleAuthGame_ResetsHistory(t *testing.T) {
	gs := newTestState()
	gs.History = append(gs.History, models.RoundResult{Winner: 0})

	data := json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`)
	gs.HandleEvent("authGame", data)

	assert.Empty(t, gs.History)
}

func TestHandleNewRound(t *testing.T) {
	gs := newTestState()
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"chang":1,"ju":2,"ben":1,"scores":[30000,25000,20000,25000]}`)
	assert.True(t, gs.HandleEvent("newRound", data))

	assert.Equal(t, 1, gs.Round.Chang)
	assert.Equal(t, 2, gs.Round.Ju)
	assert.Equal(t, 1, gs.Round.Ben)
	assert.Equal(t, 30000, gs.Players[0].Score)
	assert.Equal(t, 1, gs.Players[0].Rank)
	assert.Equal(t, 20000, gs.Players[2].Score)
	assert.Equal(t, 4, gs.Players[2].Rank)
}

func TestHandleHule(t *testing.T) {
	gs := newTestState()
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[33000,17000,25000,25000],"deltaScores":[8000,-8000,0,0],"winner":0}`)
	assert.True(t, gs.HandleEvent("hule", data))

	assert.Len(t, gs.History, 1)
	assert.Equal(t, 0, gs.History[0].Winner)
	assert.Equal(t, 33000, gs.Players[0].Score)
	assert.Equal(t, 1, gs.Players[0].Rank)
	assert.Equal(t, 17000, gs.Players[1].Score)
	assert.Equal(t, 4, gs.Players[1].Rank)

	assert.Len(t, gs.History[0].ScoreChanges, 4)
	assert.Equal(t, 8000, gs.History[0].ScoreChanges[0].Delta)
	assert.Equal(t, -8000, gs.History[0].ScoreChanges[1].Delta)
}

func TestHandleNoTile(t *testing.T) {
	gs := newTestState()
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[26000,24000,25000,25000],"deltaScores":[1000,-1000,0,0]}`)
	assert.True(t, gs.HandleEvent("noTile", data))

	assert.Len(t, gs.History, 1)
	assert.Equal(t, -1, gs.History[0].Winner)
	assert.Equal(t, 26000, gs.Players[0].Score)
}

func TestHandleLiuju(t *testing.T) {
	gs := newTestState()
	assert.True(t, gs.HandleEvent("liuju", json.RawMessage(`{}`)))
	assert.Len(t, gs.History, 1)
	assert.Equal(t, -1, gs.History[0].Winner)
}

func TestHandleGameEnd(t *testing.T) {
	gs := newTestState()
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[35000,15000,25000,25000],"ranks":[1],"deltaPt":45}`)
	assert.True(t, gs.HandleEvent("gameEnd", data))

	assert.False(t, gs.InGame)
	assert.Len(t, gs.Session, 1)
	assert.Equal(t, 1, gs.Session[0].Rank)
	assert.Equal(t, 35000, gs.Session[0].Score)
	assert.Equal(t, 45, gs.Session[0].DeltaPt)
}

func TestHandleGameEnd_DuplicateIgnored(t *testing.T) {
	gs := newTestState()
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[35000,15000,25000,25000],"ranks":[1],"deltaPt":45}`)
	gs.HandleEvent("gameEnd", data)

	// 2回目は InGame=false なので無視される
	assert.False(t, gs.HandleEvent("gameEnd", data))
	assert.Len(t, gs.Session, 1)
}

func TestHandleRankPoint(t *testing.T) {
	gs := newTestState()
	data := json.RawMessage(`{"currentPt":1500,"targetPt":2000,"rankName":"雀傑3"}`)
	assert.True(t, gs.HandleEvent("rankPoint", data))

	assert.Equal(t, 1500, gs.RankPoint.CurrentPt)
	assert.Equal(t, 2000, gs.RankPoint.TargetPt)
	assert.Equal(t, "雀傑3", gs.RankPoint.RankName)
}

func TestUpdateRanks(t *testing.T) {
	gs := newTestState()
	gs.Players = models.Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 30000},
		{Seat: 2, Score: 20000},
		{Seat: 3, Score: 25000},
	}
	gs.Players.UpdateRanks()

	assert.Equal(t, 2, gs.Players[0].Rank) // 25000, seat 0
	assert.Equal(t, 1, gs.Players[1].Rank) // 30000
	assert.Equal(t, 4, gs.Players[2].Rank) // 20000
	assert.Equal(t, 3, gs.Players[3].Rank) // 25000, seat 3
}

func TestUpdateRanks_AllSameScore(t *testing.T) {
	gs := newTestState()
	gs.Players = models.Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 25000},
		{Seat: 2, Score: 25000},
		{Seat: 3, Score: 25000},
	}
	gs.Players.UpdateRanks()

	assert.Equal(t, 1, gs.Players[0].Rank)
	assert.Equal(t, 2, gs.Players[1].Rank)
	assert.Equal(t, 3, gs.Players[2].Rank)
	assert.Equal(t, 4, gs.Players[3].Rank)
}

func TestClearSession(t *testing.T) {
	gs := newTestState()
	gs.Session = append(gs.Session, models.SessionResult{Rank: 1})
	gs.ClearSession()
	assert.Empty(t, gs.Session)
}

func TestJSON(t *testing.T) {
	gs := newTestState()
	data, err := gs.JSON()
	assert.NoError(t, err)

	var result map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &result))
	assert.Equal(t, false, result["inGame"])
	assert.NotNil(t, result["players"])
}

func TestFullGameFlow(t *testing.T) {
	gs := newTestState()

	// 対局開始
	gs.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))
	assert.True(t, gs.InGame)

	// 東1局
	gs.HandleEvent("newRound", json.RawMessage(`{"chang":0,"ju":0,"ben":0,"scores":[25000,25000,25000,25000]}`))
	assert.Equal(t, 0, gs.Round.Chang)

	// 和了
	gs.HandleEvent("hule", json.RawMessage(`{"scores":[33000,17000,25000,25000],"deltaScores":[8000,-8000,0,0],"winner":0}`))
	assert.Equal(t, 33000, gs.Players[0].Score)
	assert.Len(t, gs.History, 1)

	// 東2局
	gs.HandleEvent("newRound", json.RawMessage(`{"chang":0,"ju":1,"ben":0,"scores":[33000,17000,25000,25000]}`))
	assert.Equal(t, 1, gs.Round.Ju)

	// 流局
	gs.HandleEvent("noTile", json.RawMessage(`{"scores":[34000,16000,25000,25000],"deltaScores":[1000,-1000,0,0]}`))
	assert.Len(t, gs.History, 2)

	// 対局終了
	gs.HandleEvent("gameEnd", json.RawMessage(`{"scores":[34000,16000,25000,25000],"ranks":[1],"deltaPt":50}`))
	assert.False(t, gs.InGame)
	assert.Len(t, gs.Session, 1)
	assert.Equal(t, 1, gs.Session[0].Rank)
}
