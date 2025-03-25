package user

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

func (u User) isTrackingEnabled() bool {
	return u.TrackingStatus == TrackingStatusEnabled
}

func (u User) isTrackingDisabled() bool {
	return u.TrackingStatus == TrackingStatusDisabled
}
