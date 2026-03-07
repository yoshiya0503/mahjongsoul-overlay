package store

import (
	"encoding/json"
	"os"

	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/models"
)

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) SaveSession(session []models.SessionResult) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *FileStore) LoadSession() ([]models.SessionResult, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var session []models.SessionResult
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return session, nil
}
