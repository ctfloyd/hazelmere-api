package api

type TrackingStatus string

const (
	TrackingStatusEnabled  TrackingStatus = "ENABLED"
	TrackingStatusDisabled TrackingStatus = "DISABLED"
)

func TrackingStatusFromValue(value string) TrackingStatus {
	if value == string(TrackingStatusEnabled) {
		return TrackingStatusEnabled
	}

	if value == string(TrackingStatusDisabled) {
		return TrackingStatusDisabled
	}

	return TrackingStatusEnabled
}

type User struct {
	Id             string         `json:"id"`
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
}

type GetAllUsersResponse struct {
	Users []User `json:"users"`
}

type GetUserByIdResponse struct {
	User User `json:"user"`
}

type CreateUserRequest struct {
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
}
type CreateUserResponse struct {
	User User `json:"user"`
}

type UpdateUserRequest struct {
	Id             string         `json:"id"`
	RunescapeName  string         `json:"runescapeName"`
	TrackingStatus TrackingStatus `json:"trackingStatus"`
}
type UpdateUserResponse struct {
	User User `json:"user"`
}
