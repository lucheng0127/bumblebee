package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ConsulConfig struct {
	Url string `yaml:"url"`
}

type PlatformConfig struct {
	Zone   string `yaml:"zone"`
	Master bool   `yaml:"master"`
	Port   int    `yaml:"port"`
}

type ApiServerConfig struct {
	Platform *PlatformConfig `yaml:"platform"`
	Consul   *ConsulConfig   `yaml:"consul"`
}

func newApiServerConfig() *ApiServerConfig {
	cfg := new(ApiServerConfig)
	cfg.Platform = new(PlatformConfig)
	return cfg
}

func LoadFromFile(cfgFile string) (*ApiServerConfig, error) {
	f, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := newApiServerConfig()
	if err := yaml.Unmarshal(f, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
