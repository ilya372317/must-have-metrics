package config

import (
	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host string `env:"ADDRESS"`
}

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func (a *AgentConfig) Init() error {
	err := env.Parse(a)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerConfig) Init() error {
	err := env.Parse(s)
	if err != nil {
		return err
	}
	return nil
}
