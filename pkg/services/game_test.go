package services

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

func newTestService() *Game {
	return NewGame(nil)
}

func TestHandleEvent_UnknownType(t *testing.T) {
	svc := newTestService()
	assert.False(t, svc.HandleEvent("unknown", json.RawMessage(`{}`)))
}

func TestHandleAuthGame(t *testing.T) {
	svc := newTestService()
	data := json.RawMessage(`{"players":[{"name":"Alice","character":"1"},{"name":"Bob","character":"2"},{"name":"Carol","character":"3"},{"name":"Dave","character":"4"}]}`)

	assert.True(t, svc.HandleEvent("authGame", data))
	assert.True(t, svc.state.InGame)
	assert.Len(t, svc.state.Players, 4)
	assert.Equal(t, "Alice", svc.state.Players[0].Name)
	assert.Equal(t, "Dave", svc.state.Players[3].Name)
	assert.Equal(t, 25000, svc.state.Players[0].Score)
	assert.Equal(t, 0, svc.state.Players[0].Seat)
	assert.Equal(t, 1, svc.state.Players[0].Rank)
}

func TestHandleAuthGame_InvalidJSON(t *testing.T) {
	svc := newTestService()
	assert.False(t, svc.HandleEvent("authGame", json.RawMessage(`invalid`)))
}

func TestHandleAuthGame_ResetsHistory(t *testing.T) {
	svc := newTestService()
	svc.state.History = append(svc.state.History, models.RoundResult{Winner: 0})

	data := json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`)
	svc.HandleEvent("authGame", data)

	assert.Empty(t, svc.state.History)
}

func TestHandleNewRound(t *testing.T) {
	svc := newTestService()
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"chang":1,"ju":2,"ben":1,"scores":[30000,25000,20000,25000]}`)
	assert.True(t, svc.HandleEvent("newRound", data))

	assert.Equal(t, 1, svc.state.Round.Chang)
	assert.Equal(t, 2, svc.state.Round.Ju)
	assert.Equal(t, 1, svc.state.Round.Ben)
	assert.Equal(t, 30000, svc.state.Players[0].Score)
	assert.Equal(t, 1, svc.state.Players[0].Rank)
	assert.Equal(t, 20000, svc.state.Players[2].Score)
	assert.Equal(t, 4, svc.state.Players[2].Rank)
}

func TestHandleHule(t *testing.T) {
	svc := newTestService()
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[33000,17000,25000,25000],"deltaScores":[8000,-8000,0,0],"winner":0}`)
	assert.True(t, svc.HandleEvent("hule", data))

	assert.Len(t, svc.state.History, 1)
	assert.Equal(t, 0, svc.state.History[0].Winner)
	assert.Equal(t, 33000, svc.state.Players[0].Score)
	assert.Equal(t, 1, svc.state.Players[0].Rank)
	assert.Equal(t, 17000, svc.state.Players[1].Score)
	assert.Equal(t, 4, svc.state.Players[1].Rank)

	assert.Len(t, svc.state.History[0].ScoreChanges, 4)
	assert.Equal(t, 8000, svc.state.History[0].ScoreChanges[0].Delta)
	assert.Equal(t, -8000, svc.state.History[0].ScoreChanges[1].Delta)
}

func TestHandleNoTile(t *testing.T) {
	svc := newTestService()
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[26000,24000,25000,25000],"deltaScores":[1000,-1000,0,0]}`)
	assert.True(t, svc.HandleEvent("noTile", data))

	assert.Len(t, svc.state.History, 1)
	assert.Equal(t, -1, svc.state.History[0].Winner)
	assert.Equal(t, 26000, svc.state.Players[0].Score)
}

func TestHandleLiuju(t *testing.T) {
	svc := newTestService()
	assert.True(t, svc.HandleEvent("liuju", json.RawMessage(`{}`)))
	assert.Len(t, svc.state.History, 1)
	assert.Equal(t, -1, svc.state.History[0].Winner)
}

func TestHandleGameEnd(t *testing.T) {
	svc := newTestService()
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[35000,15000,25000,25000],"ranks":[1],"deltaPt":45}`)
	assert.True(t, svc.HandleEvent("gameEnd", data))

	assert.False(t, svc.state.InGame)
	assert.Len(t, svc.state.Session, 1)
	assert.Equal(t, 1, svc.state.Session[0].Rank)
	assert.Equal(t, 35000, svc.state.Session[0].Score)
	assert.Equal(t, 45, svc.state.Session[0].DeltaPt)
}

func TestHandleGameEnd_DuplicateIgnored(t *testing.T) {
	svc := newTestService()
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))

	data := json.RawMessage(`{"scores":[35000,15000,25000,25000],"ranks":[1],"deltaPt":45}`)
	svc.HandleEvent("gameEnd", data)

	// 2回目は InGame=false なので無視される
	assert.False(t, svc.HandleEvent("gameEnd", data))
	assert.Len(t, svc.state.Session, 1)
}

func TestHandleRankPoint(t *testing.T) {
	svc := newTestService()
	data := json.RawMessage(`{"currentPt":1500,"targetPt":2000,"rankName":"雀傑3"}`)
	assert.True(t, svc.HandleEvent("rankPoint", data))

	assert.Equal(t, 1500, svc.state.RankPoint.CurrentPt)
	assert.Equal(t, 2000, svc.state.RankPoint.TargetPt)
	assert.Equal(t, "雀傑3", svc.state.RankPoint.RankName)
}


func TestClearSession(t *testing.T) {
	svc := newTestService()
	svc.state.Session = append(svc.state.Session, models.SessionResult{Rank: 1})
	svc.ClearSession()
	assert.Empty(t, svc.state.Session)
}

func TestJSON(t *testing.T) {
	svc := newTestService()
	data, err := svc.JSON()
	assert.NoError(t, err)

	var result map[string]any
	assert.NoError(t, json.Unmarshal(data, &result))
	assert.Equal(t, false, result["inGame"])
	assert.NotNil(t, result["players"])
}

func TestFullGameFlow(t *testing.T) {
	svc := newTestService()

	// 対局開始
	svc.HandleEvent("authGame", json.RawMessage(`{"players":[{"name":"A","character":""},{"name":"B","character":""},{"name":"C","character":""},{"name":"D","character":""}]}`))
	assert.True(t, svc.state.InGame)

	// 東1局
	svc.HandleEvent("newRound", json.RawMessage(`{"chang":0,"ju":0,"ben":0,"scores":[25000,25000,25000,25000]}`))
	assert.Equal(t, 0, svc.state.Round.Chang)

	// 和了
	svc.HandleEvent("hule", json.RawMessage(`{"scores":[33000,17000,25000,25000],"deltaScores":[8000,-8000,0,0],"winner":0}`))
	assert.Equal(t, 33000, svc.state.Players[0].Score)
	assert.Len(t, svc.state.History, 1)

	// 東2局
	svc.HandleEvent("newRound", json.RawMessage(`{"chang":0,"ju":1,"ben":0,"scores":[33000,17000,25000,25000]}`))
	assert.Equal(t, 1, svc.state.Round.Ju)

	// 流局
	svc.HandleEvent("noTile", json.RawMessage(`{"scores":[34000,16000,25000,25000],"deltaScores":[1000,-1000,0,0]}`))
	assert.Len(t, svc.state.History, 2)

	// 対局終了
	svc.HandleEvent("gameEnd", json.RawMessage(`{"scores":[34000,16000,25000,25000],"ranks":[1],"deltaPt":50}`))
	assert.False(t, svc.state.InGame)
	assert.Len(t, svc.state.Session, 1)
	assert.Equal(t, 1, svc.state.Session[0].Rank)
}
