package config

type Config struct {
	// "host:port" format
	ServerAddresses      []string
	Replication          int
	PhysicalVolumeNumber int
	NodeID               int
	SequenceNumber       int
}

var SystemConfig Config

func init() {
	SystemConfig.ServerAddresses = []string{"localhost:4000"}
	SystemConfig.Replication = 3
	SystemConfig.PhysicalVolumeNumber = 3
}
