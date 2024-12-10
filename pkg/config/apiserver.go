package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type NatsConfig struct {
	Url string `yaml:"url"`
}

type TraceConfig struct {
	Addr   string `yaml:"rpcAddr"`
	Enable bool   `yaml:"enable"`
}

type ConsulConfig struct {
	Url string `yaml:"url"`
}

type AdminUserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PlatformConfig struct {
	Zone      string           `yaml:"zone"`
	Master    bool             `yaml:"master"`
	Port      int              `yaml:"port"`
	Secret    string           `yaml:"secret"`
	AdminUser *AdminUserConfig `yaml:"adminUser"`
}

type ApiServerConfig struct {
	Platform *PlatformConfig `yaml:"platform"`
	Consul   *ConsulConfig   `yaml:"consul"`
	Trace    *TraceConfig    `yaml:"trace"`
	Nats     *NatsConfig     `yaml:"nats"`
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
