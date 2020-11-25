package master

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	WebApp           string `toml:"WebApp"`
	WebIndex         string `toml:"WebIndex"`
	Service          string `toml:"Service"`
	SliceSize        int    `toml:"SliceSize"`
	SaveFileMultiple bool   `toml:"SaveFileMultiple"`
}

var config *Config

func LoadConfig(path string) *Config {
	conf := &Config{}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		panic(err)
	}
	conf.SaveFileMultiple = true
	conf.SliceSize = 2
	config = conf
	return conf
}

func getConfig() *Config {
	return config
}
