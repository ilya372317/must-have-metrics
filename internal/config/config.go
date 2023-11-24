package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
	"github.com/ilya372317/must-have-metrics/internal/config/params"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
	"github.com/joho/godotenv"
)

var serverParams = map[string]Parameter{
	"host":           &params.HostConfig{},
	"store_interval": &params.StoreIntervalConfig{},
	"store_path":     &params.StoreFilePathConfig{},
	"restore":        &params.RestoreConfig{},
}

var agentParams = map[string]Parameter{
	"host":            &params.HostConfig{},
	"poll_interval":   &params.PollIntervalConfig{},
	"report_interval": &params.ReportIntervalConfig{},
}

type Configuration interface {
	GetParameters() map[string]Parameter
}

type Parameter interface {
	SetFlag(*string)
	GetFlag() *string
	GetEnv() string
	SetEnv(string)
	SetValue(string)
	GetValue() string
	GetFlagName() string
	GetDefaultFlagValue() string
	GetFlagDescription() string
}

func initConfiguration(config Configuration) error {
	if err := godotenv.Load(".env-server"); err != nil {
		logger.Get().Warnf("failed load .env-server file: %v", err)
	}

	for _, paramConfig := range config.GetParameters() {
		if err := env.Parse(paramConfig); err != nil {
			return fmt.Errorf("failed parse server parameters: %w", err)
		}
		paramConfig.SetFlag(
			flag.String(
				paramConfig.GetFlagName(),
				paramConfig.GetDefaultFlagValue(),
				paramConfig.GetFlagDescription(),
			),
		)
	}

	flag.Parse()

	for _, paramConfig := range config.GetParameters() {
		if paramConfig.GetEnv() == "" {
			paramConfig.SetValue(*paramConfig.GetFlag())
		} else {
			paramConfig.SetValue(paramConfig.GetEnv())
		}
	}

	return nil
}
