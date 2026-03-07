package models

type Player struct {
	Seat      int    `json:"seat"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Rank      int    `json:"rank"`
	Character string `json:"character"`
}

type Players []Player

func (ps Players) UpdateScores(scores []int) {
	if len(scores) != len(ps) {
		return
	}
	for i := range ps {
		ps[i].Score = scores[i]
	}
	ps.UpdateRanks()
}

func (ps Players) UpdateRanks() {
	for i := range ps {
		rank := 1
		for j := range ps {
			if i != j && (ps[j].Score > ps[i].Score ||
				(ps[j].Score == ps[i].Score && ps[j].Seat < ps[i].Seat)) {
				rank++
			}
		}
		ps[i].Rank = rank
	}
}
