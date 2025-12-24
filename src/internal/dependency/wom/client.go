package wom

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
)

const requestsPerMinute = 7
const maxLimit = 50

type Client struct {
	h             *hz_client.HttpClient
	nextRequestAt time.Time
}

func NewClient(logger hz_logger.Logger) *Client {
	wom := hz_client.NewHttpClient(hz_client.HttpClientConfig{
		Host:           "https://api.wiseoldman.net",
		TimeoutMs:      5000,
		Retries:        0,
		RetryWaitMs:    0,
		RetryMaxWaitMs: 0,
	}, func(str string) { logger.Error(context.TODO(), str) })

	return &Client{
		h: wom,
	}
}

func (c *Client) GetPlayerDetails(username string) (PlayerDetails, error) {
	url := fmt.Sprintf("%s/v2/players/%s", c.h.GetHost(), username)
	attempts := 0
	for attempts < 2 {
		c.modulate()
		var response PlayerDetails
		if err := c.h.Get(url, &response); err != nil {
			time.Sleep(10 * time.Second)
			attempts++
			continue
		}
		return response, nil
	}
	return PlayerDetails{}, errors.New("ran out of attempts")
}

func (c *Client) GetPlayerSnapshots(username string, startTime time.Time, endTime time.Time) ([]Snapshot, error) {
	snapshots := make([]Snapshot, 0)

	done := false
	page := 0
	errorCount := 0
	for !done {
		slog.Info("page", slog.Int("page", page))
		c.modulate()
		url := fmt.Sprintf("%s/v2/players/%s/snapshots?startDate=%s&endDate=%s&limit=%d&offset=%d",
			c.h.GetHost(),
			username,
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"),
			maxLimit,
			page*maxLimit,
		)

		var response []Snapshot

		if err := c.h.Get(url, &response); err != nil {
			errorCount++
			if errorCount >= 5 {
				return nil, err
			}
			time.Sleep(10 * time.Second)
			continue
		}

		snapshots = append(snapshots, response...)

		page++
		errorCount = 0
		if len(response) != 50 {
			done = true
		}
	}

	return snapshots, nil
}

func (c *Client) getRequestOffset() time.Duration {
	return (60 / requestsPerMinute) * time.Second
}

func (c *Client) modulate() {
	now := time.Now()
	if time.Now().Before(c.nextRequestAt) {
		dur := c.nextRequestAt.Sub(now)
		time.Sleep(dur)
	}
	c.nextRequestAt = time.Now().Add(c.getRequestOffset())
}
