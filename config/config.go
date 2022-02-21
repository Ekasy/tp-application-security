package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port     string `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	}
}

func NewConfig(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	yaml_config := &Config{}
	err = yaml.Unmarshal(bytes, yaml_config)
	if err != nil {
		return nil, err
	}

	return yaml_config, nil
}
