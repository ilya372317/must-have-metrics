package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	defaultAgentRateLimitValue      = 1
	defaultAgentPollIntervalValue   = 2
	defaultAgentReportIntervalValue = 10
	defaultAgentAddressValue        = "localhost:8080"
	defaultAgentSecretKeyValue      = ""
	defaultAgentCryptoKeyValue      = ""
	defaultAgentConfigValue         = ""

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
	ConfigPath     string `env:"CONFIG"`
	PollInterval   uint   `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	ReportInterval uint   `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	RateLimit      uint   `env:"RATE_LIMIT" json:"rate_limit,omitempty"`
}

// NewAgent constructor for AgentConfig.
func NewAgent() (*AgentConfig, error) {
	agentConfig := &AgentConfig{}
	agentConfig.parseFlags()
	if err := env.Parse(agentConfig); err != nil {
		return nil, fmt.Errorf("failed parse agent flags: %w", err)
	}
	if err := agentConfig.parseFromFile(); err != nil {
		return nil, fmt.Errorf("failed create agent config: %w", err)
	}

	return agentConfig, nil
}

func (c *AgentConfig) parseFlags() {
	flag.StringVar(
		&c.Host, "a",
		defaultAgentAddressValue, "address where server will listen requests",
	)
	flag.UintVar(&c.PollInterval, "p", defaultAgentPollIntervalValue, "interval agent collect metrics")
	flag.UintVar(&c.ReportInterval, "r", defaultAgentReportIntervalValue, "interval agent send metrics on server")
	flag.StringVar(&c.SecretKey, "k", defaultAgentSecretKeyValue, "secret key for sign")
	flag.UintVar(&c.RateLimit, "l", defaultAgentRateLimitValue, "limit of simultaneously requests to server")
	flag.StringVar(&c.CryptoKey, "crypto-key", defaultAgentCryptoKeyValue, "public crypto key for cipher transferred data")
	flag.StringVar(&c.ConfigPath, "c", defaultAgentConfigValue, "file path to json configuration file")
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

	tempConfig := AgentConfig{
		Host:           defaultAgentAddressValue,
		SecretKey:      defaultAgentSecretKeyValue,
		CryptoKey:      defaultAgentCryptoKeyValue,
		ConfigPath:     defaultAgentConfigValue,
		PollInterval:   defaultAgentPollIntervalValue,
		ReportInterval: defaultAgentReportIntervalValue,
		RateLimit:      defaultAgentRateLimitValue,
	}

	if err = json.Unmarshal(fileContent, &tempConfig); err != nil {
		return fmt.Errorf("invalid config file content: %w", err)
	}

	if c.Host == defaultAgentAddressValue || c.Host == nullStringValue {
		c.Host = tempConfig.Host
	}
	if c.SecretKey == defaultAgentSecretKeyValue {
		c.SecretKey = tempConfig.SecretKey
	}
	if c.CryptoKey == defaultAgentCryptoKeyValue {
		c.CryptoKey = tempConfig.CryptoKey
	}
	if c.PollInterval == defaultAgentPollIntervalValue || c.PollInterval == nullIntValue {
		c.PollInterval = tempConfig.PollInterval
	}
	if c.ReportInterval == defaultAgentReportIntervalValue || c.ReportInterval == nullIntValue {
		c.ReportInterval = tempConfig.ReportInterval
	}
	if c.RateLimit == defaultAgentRateLimitValue || c.RateLimit == nullIntValue {
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
