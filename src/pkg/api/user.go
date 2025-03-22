package api

type TrackingStatus string

const (
	TrackingStatusEnabled  TrackingStatus = "ENABLED"
	TrackingStatusDisabled TrackingStatus = "DISABLED"
)

type User struct {
	Id             string
	RunescapeName  string
	TrackingStatus TrackingStatus
}
