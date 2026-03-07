package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGameState(t *testing.T) {
	gs := NewGameState()

	assert.False(t, gs.InGame)
	assert.NotNil(t, gs.Players)
	assert.Empty(t, gs.Players)
	assert.NotNil(t, gs.History)
	assert.Empty(t, gs.History)
	assert.NotNil(t, gs.Session)
	assert.Empty(t, gs.Session)
}
