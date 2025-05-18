package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server Server `json:"server" yaml:"server"`
}

type Server struct {
	Port              int           `json:"port"                yaml:"port"`
	ReadTimeout       time.Duration `json:"read_timeout"        yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"       yaml:"write_timeout"`
}

func GetConfigFromFile(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
