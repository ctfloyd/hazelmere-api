package main

import (
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-api/src/pkg/client"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
)

func main() {
	type skill struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Rank  int    `json:"rank"`
		Level int    `json:"level"`
		Xp    int    `json:"xp"`
	}

	type activity struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Rank  int    `json:"rank"`
		Score int    `json:"score"`
	}

	type hiscore struct {
		Skills     []skill    `json:"skills"`
		Activities []activity `json:"activities"`
	}

	req, err := http.NewRequest("GET", "https://secure.runescape.com/m=hiscore_oldschool/index_lite.json?player=st%20jamie", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	var hs hiscore
	hiBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	err = jsoniter.Unmarshal(hiBytes, &hs)
	if err != nil {
		panic(err)
	}

	snapshot := api.HiscoreSnapshot{
		UserId:     uuid.New().String(),
		Timestamp:  time.Now(),
		Skills:     make([]api.SkillSnapshot, 0),
		Bosses:     make([]api.BossSnapshot, 0),
		Activities: make([]api.ActivitySnapshot, 0),
	}

	for _, skill := range hs.Skills {
		snapshot.Skills = append(snapshot.Skills, api.SkillSnapshot{
			ActivityType: api.ActivityTypeFromValue(sanitizeNameUpper(skill.Name)),
			Name:         skill.Name,
			Level:        skill.Level,
			Experience:   skill.Xp,
			Rank:         skill.Rank,
		})
	}

	for _, activity := range hs.Activities {
		at := api.ActivityTypeFromValue(sanitizeNameUpper(activity.Name))

		if slices.Contains(api.AllBossActivityTypes, at) {
			snapshot.Bosses = append(snapshot.Bosses, api.BossSnapshot{
				ActivityType: at,
				Name:         activity.Name,
				KillCount:    activity.Score,
				Rank:         activity.Rank,
			})
		} else {
			snapshot.Activities = append(snapshot.Activities, api.ActivitySnapshot{
				ActivityType: at,
				Name:         activity.Name,
				Score:        activity.Score,
				Rank:         activity.Rank,
			})
		}
	}

	config := client.HazelmereClientConfig{
		Host:      "https://api.hazelmere.xyz",
		TimeoutMs: 500,
	}
	cl := client.NewHazelmereClient(config, func(msg string) { fmt.Printf(msg) })
	hazelmere := client.NewHazelmere(cl)

	start := time.Now()
	userId := uuid.New().String()
	for i := 0; i < 100; i++ {
		snapshot.Timestamp = start.Add(-1 * time.Duration(i) * time.Hour * 24)
		snapshot.UserId = userId
		createRequest := api.CreateSnapshotRequest{
			Snapshot: snapshot,
		}

		response, err := hazelmere.Snapshot.CreateSnapshot(createRequest)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d: %v\n", i, response)
	}

}

func removeIllegalChars(str string) string {
	remove := regexp.MustCompile("['\\-_:()]")
	return remove.ReplaceAllString(str, "")
}

func sanitizeNameUpper(name string) string {
	name = removeIllegalChars(name)
	return strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
}
