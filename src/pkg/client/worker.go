package client

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
)

var ErrHiscoreTimeout = errors.Join(ErrHazelmereClient, errors.New("osrs hiscore timeout"))

type Worker struct {
	prefix string
	client *hz_client.HttpClient
}

func newWorker(client *hz_client.HttpClient) *Worker {
	mappings := map[string]error{
		api.ErrorCodeHiscoreTimeout: ErrHiscoreTimeout,
	}
	client.AddErrorMappings(mappings)
	return &Worker{
		prefix: "worker",
		client: client,
	}
}

func (worker *Worker) GenerateSnapshotOnDemand(userId string) (api.GenerateSnapshotOnDemandResponse, error) {
	var response api.GenerateSnapshotOnDemandResponse
	url := fmt.Sprintf("%s/snapshot/on-demand/%s", worker.getBaseUrl(), userId)
	err := worker.client.Get(url, &response)
	if err != nil {
		return api.GenerateSnapshotOnDemandResponse{}, err
	}
	return response, nil
}

func (worker *Worker) getBaseUrl() string {
	return fmt.Sprintf("%s/%s", worker.client.GetV1Url(), worker.prefix)
}
