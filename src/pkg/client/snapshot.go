package client

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
)

var ErrSnapshotNotFound = errors.Join(ErrHazelmereClient, errors.New("snapshot not found"))
var ErrInvalidSnapshot = errors.Join(ErrHazelmereClient, errors.New("invalid snapshot"))

type Snapshot struct {
	prefix string
	client *HazelmereClient
}

func newSnapshot(client *HazelmereClient) *Snapshot {
	mappings := map[string]error{
		api.ErrorCodeSnapshotNotFound: ErrSnapshotNotFound,
		api.ErrorCodeInvalidSnapshot:  ErrInvalidSnapshot,
	}
	client.AddErrorMappings(mappings)

	return &Snapshot{
		prefix: "snapshot",
		client: client,
	}
}

func (ss *Snapshot) GetAllSnapshotsForUser(userId string) (api.GetAllSnapshotsForUser, error) {
	url := fmt.Sprintf("%s/%s", ss.getBaseUrl(), userId)
	var response api.GetAllSnapshotsForUser
	err := ss.client.Get(url, &response)
	if err != nil {
		return api.GetAllSnapshotsForUser{}, err
	}
	return response, nil
}

func (ss *Snapshot) GetSnapshotForUserNearestTimestamp(userId string, epochMillis int64) (api.GetSnapshotNearestTimestampResponse, error) {
	url := fmt.Sprintf("%s/%s/nearest/%d", ss.getBaseUrl(), userId, epochMillis)
	var response api.GetSnapshotNearestTimestampResponse
	err := ss.client.Get(url, &response)
	if err != nil {
		return api.GetSnapshotNearestTimestampResponse{}, err
	}
	return response, nil
}

func (ss *Snapshot) CreateSnapshot(request api.CreateSnapshotRequest) (api.CreateSnapshotResponse, error) {
	url := ss.getBaseUrl()
	var response api.CreateSnapshotResponse
	err := ss.client.Post(url, request, &response)
	if err != nil {
		return api.CreateSnapshotResponse{}, err
	}
	return response, nil
}

func (ss *Snapshot) getBaseUrl() string {
	return fmt.Sprintf("%s/%s", ss.client.GetV1Url(), ss.prefix)
}
