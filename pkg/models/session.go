package models

import "time"

type SessionResult struct {
	Rank      int       `json:"rank"`
	Score     int       `json:"score"`
	DeltaPt   int       `json:"deltaPt"`
	Timestamp time.Time `json:"timestamp"`
}
