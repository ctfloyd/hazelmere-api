package client

import (
	"errors"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
)

var ErrHazelmereClient = errors.New("generic hazelmere client error")
var ErrHazelmereUnauthorized = errors.Join(ErrHazelmereClient, errors.New("unauthorized"))
var ErrIllegalArgument = errors.Join(ErrHazelmereClient, errors.New("illegal argument"))

type Hazelmere struct {
	Snapshot *Snapshot
	User     *User
	Worker   *Worker
	Delta    *Delta
	Config   HazelmereConfig
}

type HazelmereConfig struct {
	Token              string
	CallingApplication string
}

func NewHazelmere(client *hz_client.HttpClient, config HazelmereConfig) (*Hazelmere, error) {
	if client == nil {
		return nil, errors.Join(ErrIllegalArgument, errors.New("client is nil"))
	}

	mappings := map[string]error{
		api.ErrorCodeUnauthorized: ErrHazelmereUnauthorized,
	}
	client.AddErrorMappings(mappings)

	return &Hazelmere{
		Snapshot: newSnapshot(client, config),
		User:     newUser(client, config),
		Worker:   newWorker(client, config),
		Delta:    newDelta(client, config),
		Config:   config,
	}, nil
}
