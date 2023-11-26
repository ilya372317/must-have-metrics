package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
	"github.com/ilya372317/must-have-metrics/internal/config/params"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
	"github.com/joho/godotenv"
)

const (
	Host           = "host"
	StoreInterval  = "store_interval"
	StorePath      = "store_path"
	Restore        = "restore"
	PollInterval   = "poll_interval"
	ReportInterval = "report_interval"
)

var serverParams = map[string]Parameter{
	Host:          &params.HostConfig{},
	StoreInterval: &params.StoreIntervalConfig{},
	StorePath:     &params.StoreFilePathConfig{},
	Restore:       &params.RestoreConfig{},
}

var agentParams = map[string]Parameter{
	Host:           &params.HostConfig{},
	PollInterval:   &params.PollIntervalConfig{},
	ReportInterval: &params.ReportIntervalConfig{},
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

func initConfiguration(config Configuration, isServer bool) error {
	var envFileName string
	if isServer {
		envFileName = ".env-server"
	} else {
		envFileName = ".env-agent"
	}

	if err := godotenv.Load(envFileName); err != nil {
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
