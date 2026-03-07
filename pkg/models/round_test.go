package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRoundResult(t *testing.T) {
	round := RoundInfo{Chang: 0, Ju: 0, Ben: 1}
	scores := []int{33000, 17000, 25000, 25000}
	deltas := []int{8000, -8000, 0, 0}

	result := NewRoundResult(round, scores, deltas, 0)

	assert.Equal(t, round, result.Round)
	assert.Equal(t, scores, result.FinalScores)
	assert.Equal(t, 0, result.Winner)
	assert.Len(t, result.ScoreChanges, 4)
	assert.Equal(t, ScoreChange{Seat: 0, Delta: 8000}, result.ScoreChanges[0])
	assert.Equal(t, ScoreChange{Seat: 1, Delta: -8000}, result.ScoreChanges[1])
	assert.False(t, result.Timestamp.IsZero())
}

func TestNewRoundResult_NoTile(t *testing.T) {
	result := NewRoundResult(RoundInfo{}, []int{25000, 25000, 25000, 25000}, []int{0, 0, 0, 0}, -1)
	assert.Equal(t, -1, result.Winner)
}
