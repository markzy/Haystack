package config

type Config struct {
	ServerAddresses []string
	SequenceNumber  int
}

var SystemConfig Config

func init() {
	SystemConfig.ServerAddresses = []string{"localhost:4000"}
}
