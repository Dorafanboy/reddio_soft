package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DelayConfig struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

type RetrySettings struct {
	MaxRetries    int `yaml:"max_retries"`
	RetryDelaySec int `yaml:"retry_delay_sec"`
}

type OnchainDelayConfig struct {
	BetweenAccounts DelayConfig `yaml:"between_accounts"`
	BetweenModules  DelayConfig `yaml:"between_modules"`
}

type EthBridgeConfig struct {
	MinEthAmount   string   `yaml:"min_eth_amount"`
	MaxEthAmount   string   `yaml:"max_eth_amount"`
	MinDecimals    int      `yaml:"min_decimals"`
	MaxDecimals    int      `yaml:"max_decimals"`
	EnabledModules []string `yaml:"enabled_modules"`
}

type Config struct {
	DelayBeforeLoginTwitter  DelayConfig        `yaml:"delay_before_login_twitter"`
	DelayBeforeLogin         DelayConfig        `yaml:"delay_before_login"`
	DelayBeforeDaily         DelayConfig        `yaml:"delay_before_daily"`
	DelayBeforeRepost        DelayConfig        `yaml:"delay_before_repost"`
	DelayBetweenAccs         DelayConfig        `yaml:"delay_between_accs"`
	DelayBetweenAccsIfCsv    DelayConfig        `yaml:"delay_between_accs_if_csv"`
	DelayBetweenDailyModules DelayConfig        `yaml:"between_daily_modules"`
	Mode                     string             `yaml:"mode"`
	IsShuffle                bool               `yaml:"is_shuffle"`
	RetrySettings            RetrySettings      `yaml:"retry_settings"`
	OnchainDelays            OnchainDelayConfig `yaml:"onchain_delays"`
	EthBridge                EthBridgeConfig    `yaml:"eth_bridge"`
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
