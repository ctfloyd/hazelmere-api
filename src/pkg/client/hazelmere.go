package client

type Hazelmere struct {
	Snapshot *Snapshot
	User     *User
}

func NewHazelmere(client *HazelmereClient) *Hazelmere {
	return &Hazelmere{
		Snapshot: newSnapshot(client),
		User:     newUser(client),
	}
}
