package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	defaultRateLimitValue      = 1
	defaultPollIntervalValue   = 2
	defaultReportIntervalValue = 10
	defaultAddressValue        = "localhost:8080"
	defaultSecretKeyValue      = ""
	defaultCryptoKeyValue      = ""
	defaultConfigValue         = ""

	nullStringValue = ""
	nullIntValue    = 0
)

// AgentConfig agent configs.
//
// Adding new filed steps:
// 1. Add new field to this struct.
// 2. Fill env and json struct tags.
// 3. Add parsing new field in parseFlags method.
// 4. Add parsing new field in parseFromFileMethod.
//
// Note: for default config values use constants.
type AgentConfig struct {
	Host           string `env:"ADDRESS" json:"address,omitempty"`
	SecretKey      string `env:"KEY" json:"secret_key,omitempty"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	PollInterval   uint   `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	ReportInterval uint   `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	RateLimit      uint   `env:"RATE_LIMIT" json:"rate_limit,omitempty"`
	ConfigPath     string `env:"CONFIG"`
}

// NewAgent constructor for AgentConfig.
func NewAgent() (*AgentConfig, error) {
	agentConfig := &AgentConfig{}
	agentConfig.parseFlags()
	if err := env.Parse(agentConfig); err != nil {
		return nil, fmt.Errorf("failed parse agent flags: %w", err)
	}
	if err := agentConfig.parseFromFile(); err != nil {
		return nil, fmt.Errorf("failed parse agent config from file: %w", err)
	}

	return agentConfig, nil
}

func (c *AgentConfig) parseFlags() {
	flag.StringVar(
		&c.Host, "a",
		defaultAddressValue, "address where server will listen requests",
	)
	flag.UintVar(&c.PollInterval, "p", defaultPollIntervalValue, "interval agent collect metrics")
	flag.UintVar(&c.ReportInterval, "r", defaultReportIntervalValue, "interval agent send metrics on server")
	flag.StringVar(&c.SecretKey, "k", defaultSecretKeyValue, "secret key for sign")
	flag.UintVar(&c.RateLimit, "l", defaultRateLimitValue, "limit of simultaneously requests to server")
	flag.StringVar(&c.CryptoKey, "crypto-key", defaultCryptoKeyValue, "public crypto key for cipher transferred data")
	flag.StringVar(&c.ConfigPath, "c", defaultConfigValue, "file path to json configuration file")
	flag.Parse()
}

// parseFromFile fill config fields from given file. File required be in json format.
func (c *AgentConfig) parseFromFile() error {
	if c.ConfigPath == "" {
		return nil
	}
	fileContent, err := getConfigFileContent(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed parse agent config from file: %w", err)
	}

	tempConfig := AgentConfig{}

	if err = json.Unmarshal(fileContent, &tempConfig); err != nil {
		return fmt.Errorf("invalid config file content: %w", err)
	}

	if c.Host == defaultAddressValue || c.Host == nullStringValue {
		c.Host = tempConfig.Host
	}
	if c.SecretKey == defaultSecretKeyValue {
		c.SecretKey = tempConfig.SecretKey
	}
	if c.CryptoKey == defaultCryptoKeyValue {
		c.CryptoKey = tempConfig.CryptoKey
	}
	if c.PollInterval == defaultPollIntervalValue || c.PollInterval == nullIntValue {
		c.PollInterval = tempConfig.PollInterval
	}
	if c.ReportInterval == defaultReportIntervalValue || c.ReportInterval == nullIntValue {
		c.ReportInterval = tempConfig.ReportInterval
	}
	if c.RateLimit == defaultRateLimitValue || c.RateLimit == nullIntValue {
		c.RateLimit = tempConfig.RateLimit
	}

	return nil
}

// ShouldSignData check if agent configured for sign sending data.
func (c *AgentConfig) ShouldSignData() bool {
	return c.SecretKey != ""
}

// ShouldCipherData check if agent configured for crypt sending data.
func (c *AgentConfig) ShouldCipherData() bool {
	if c.CryptoKey == "" {
		return false
	}

	if _, err := os.Stat(c.CryptoKey); os.IsNotExist(err) {
		return false
	}

	return true
}
