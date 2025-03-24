package client

type Hazelmere struct {
	Snapshot *Snapshot
}

func NewHazelmere(client *HazelmereClient) *Hazelmere {
	return &Hazelmere{
		Snapshot: newSnapshot(client),
	}
}
