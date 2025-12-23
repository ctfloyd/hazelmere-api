package client

import (
	"errors"
	"fmt"

	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
)

var ErrDeltaNotFound = errors.Join(ErrHazelmereClient, errors.New("delta not found"))

type Delta struct {
	prefix string
	client *hz_client.HttpClient
	config HazelmereConfig
}

func newDelta(client *hz_client.HttpClient, config HazelmereConfig) *Delta {
	mappings := map[string]error{
		api.ErrorCodeDeltaNotFound: ErrDeltaNotFound,
	}
	client.AddErrorMappings(mappings)

	return &Delta{
		prefix: "delta",
		client: client,
		config: config,
	}
}

func (d *Delta) GetLatestDelta(userId string) (api.GetLatestDeltaResponse, error) {
	url := fmt.Sprintf("%s/%s/latest", d.getBaseUrl(), userId)
	var response api.GetLatestDeltaResponse
	err := d.client.GetWithHeaders(url, makeHeadersFromConfig(d.config), &response)
	if err != nil {
		return api.GetLatestDeltaResponse{}, err
	}
	return response, nil
}

func (d *Delta) GetDeltaInterval(request api.GetDeltaIntervalRequest) (api.GetDeltaIntervalResponse, error) {
	url := fmt.Sprintf("%s/interval", d.getBaseUrl())
	var response api.GetDeltaIntervalResponse
	err := d.client.PostWithHeaders(url, makeHeadersFromConfig(d.config), request, &response)
	if err != nil {
		return api.GetDeltaIntervalResponse{}, err
	}
	return response, nil
}

func (d *Delta) GetDeltaSummary(request api.GetDeltaSummaryRequest) (api.GetDeltaSummaryResponse, error) {
	url := fmt.Sprintf("%s/summary", d.getBaseUrl())
	var response api.GetDeltaSummaryResponse
	err := d.client.PostWithHeaders(url, makeHeadersFromConfig(d.config), request, &response)
	if err != nil {
		return api.GetDeltaSummaryResponse{}, err
	}
	return response, nil
}

func (d *Delta) getBaseUrl() string {
	return fmt.Sprintf("%s/%s", d.client.GetV1Url(), d.prefix)
}
