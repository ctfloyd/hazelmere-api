package wom

import "time"

type PlayerDetails struct {
	RegisteredAt time.Time `json:"registeredAt"`
}

type Snapshot struct {
	CreatedAt time.Time    `json:"createdAt"`
	Data      SnapshotData `json:"data"`
}

type SnapshotData struct {
	Skills     map[string]SnapshotSkill    `json:"skills"`
	Bosses     map[string]SnapshotBoss     `json:"bosses"`
	Activities map[string]SnapshotActivity `json:"activities"`
}

type SnapshotSkill struct {
	Metric     string `json:"metric"`
	Experience int    `json:"experience"`
	Rank       int    `json:"rank"`
	Level      int    `json:"level"`
}

type SnapshotBoss struct {
	Metric string `json:"metric"`
	Kills  int    `json:"kills"`
	Rank   int    `json:"rank"`
}

type SnapshotActivity struct {
	Metric string `json:"metric"`
	Score  int    `json:"score"`
	Rank   int    `json:"rank"`
}
