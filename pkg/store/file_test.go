package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

func TestFileStore_SaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	s := NewFileStore(path)

	session := []models.SessionResult{
		{Rank: 1, Score: 35000, DeltaPt: 50, Timestamp: time.Now()},
		{Rank: 3, Score: 22000, DeltaPt: -15, Timestamp: time.Now()},
	}

	err := s.SaveSession(session)
	assert.NoError(t, err)

	loaded, err := s.LoadSession()
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, 1, loaded[0].Rank)
	assert.Equal(t, 35000, loaded[0].Score)
	assert.Equal(t, 3, loaded[1].Rank)
}

func TestFileStore_SaveEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	s := NewFileStore(path)

	err := s.SaveSession([]models.SessionResult{})
	assert.NoError(t, err)

	loaded, err := s.LoadSession()
	assert.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestFileStore_LoadNotFound(t *testing.T) {
	s := NewFileStore("/tmp/nonexistent_test_file.json")
	_, err := s.LoadSession()
	assert.True(t, os.IsNotExist(err))
}

func TestFileStore_LoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	os.WriteFile(path, []byte("not json"), 0644)

	s := NewFileStore(path)
	_, err := s.LoadSession()
	assert.Error(t, err)
}
