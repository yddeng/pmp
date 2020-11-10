package master

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	WebApp  string `toml:"WebApp"`
	Service string `toml:"Service"`
}

var config *Config

func LoadConfig(path string) *Config {
	conf := &Config{}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		panic(err)
	}
	config = conf
	return conf
}

func getConfig() *Config {
	return config
}
