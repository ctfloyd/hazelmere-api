package user

type UserData struct {
	Id             string `bson:"_id"`
	RunescapeName  string `bson:"runescapeName"`
	TrackingStatus string `bson:"trackingStatus"`
	AccountType    string `bson:"accountType"`
}
