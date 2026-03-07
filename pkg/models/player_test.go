package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateRanks(t *testing.T) {
	players := Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 30000},
		{Seat: 2, Score: 20000},
		{Seat: 3, Score: 25000},
	}
	players.UpdateRanks()

	assert.Equal(t, 2, players[0].Rank) // 25000, seat 0
	assert.Equal(t, 1, players[1].Rank) // 30000
	assert.Equal(t, 4, players[2].Rank) // 20000
	assert.Equal(t, 3, players[3].Rank) // 25000, seat 3
}

func TestUpdateRanks_AllSameScore(t *testing.T) {
	players := Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 25000},
		{Seat: 2, Score: 25000},
		{Seat: 3, Score: 25000},
	}
	players.UpdateRanks()

	assert.Equal(t, 1, players[0].Rank)
	assert.Equal(t, 2, players[1].Rank)
	assert.Equal(t, 3, players[2].Rank)
	assert.Equal(t, 4, players[3].Rank)
}

func TestUpdateScores(t *testing.T) {
	players := Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 25000},
		{Seat: 2, Score: 25000},
		{Seat: 3, Score: 25000},
	}
	players.UpdateScores([]int{30000, 20000, 25000, 25000})

	assert.Equal(t, 30000, players[0].Score)
	assert.Equal(t, 20000, players[1].Score)
	assert.Equal(t, 1, players[0].Rank)
	assert.Equal(t, 4, players[1].Rank)
}

func TestUpdateScores_LengthMismatch(t *testing.T) {
	players := Players{
		{Seat: 0, Score: 25000},
		{Seat: 1, Score: 25000},
	}
	players.UpdateScores([]int{30000})

	// スコアは変わらない
	assert.Equal(t, 25000, players[0].Score)
	assert.Equal(t, 25000, players[1].Score)
}
