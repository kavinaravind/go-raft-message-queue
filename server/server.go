package server

type Config struct {
	Address string
}

func NewServerConfig() *Config {
	return &Config{}
}
