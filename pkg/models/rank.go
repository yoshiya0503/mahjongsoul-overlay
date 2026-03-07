package models

type RankPointInfo struct {
	CurrentPt int    `json:"currentPt"`
	TargetPt  int    `json:"targetPt"`
	RankName  string `json:"rankName"`
}
