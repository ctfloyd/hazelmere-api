package client

import (
	"errors"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
)

var ErrHazelmereClient = errors.New("generic hazelmere client error")

type Hazelmere struct {
	Snapshot *Snapshot
	User     *User
	Worker   *Worker
}

func NewHazelmere(client *hz_client.HttpClient) *Hazelmere {
	return &Hazelmere{
		Snapshot: newSnapshot(client),
		User:     newUser(client),
		Worker:   newWorker(client),
	}
	t
}
