package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	SecretKey      string `env:"KEY"`
	RateLimit      uint   `env:"RATE_LIMIT"`
}

func NewAgent() (*AgentConfig, error) {
	agentConfig := &AgentConfig{}
	agentConfig.parseFlags()
	if err := env.Parse(agentConfig); err != nil {
		return nil, fmt.Errorf("failed parse agent flags: %w", err)
	}

	return agentConfig, nil
}

func (c *AgentConfig) parseFlags() {
	flag.StringVar(
		&c.Host, "a",
		"localhost:8080", "address where server will listen requests",
	)
	flag.UintVar(&c.PollInterval, "p", 2, "interval agent collect metrics")
	flag.UintVar(&c.ReportInterval, "r", 10, "interval agent send metrics on server")
	flag.StringVar(&c.SecretKey, "k", "", "Secret key for sign")
	flag.UintVar(&c.RateLimit, "l", 1, "Limit of simultaneously requests to server")
	flag.Parse()
}

func (c *AgentConfig) ShouldSignData() bool {
	return c.SecretKey != ""
}
