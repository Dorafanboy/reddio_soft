package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type DelayConfig struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

type Config struct {
	DelayBeforeLoginTwitter DelayConfig `yaml:"delay_before_login_twitter"`
	DelayBeforeLogin        DelayConfig `yaml:"delay_before_login"`
	DelayBeforeDaily        DelayConfig `yaml:"delay_before_daily"`
	DelayBeforeRepost       DelayConfig `yaml:"delay_before_repost"`
	DelayBetweenAccs        DelayConfig `yaml:"delay_between_accs"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
