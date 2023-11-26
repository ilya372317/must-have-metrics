package config

import (
	"fmt"
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var serverConfig *ServerConfig

type ServerConfig struct {
	sync.Mutex
	parameters map[string]Parameter
}

func GetServerConfig() *ServerConfig {
	if serverConfig == nil {
		logger.Get().Panicf("You forget init server configuration! Please, do it in cmd/server/main.go")
	}

	return serverConfig
}

func InitServerConfig() error {
	if serverConfig != nil {
		return nil
	}

	serverConfig = newServerConfig()
	if err := initConfiguration(serverConfig, true); err != nil {
		serverConfig = nil
		return fmt.Errorf("failed init server configuration: %w", err)
	}

	return nil
}

func (s *ServerConfig) GetValue(alias string) string {
	value := ""
	s.Mutex.Lock()
	value = s.parameters[alias].GetValue()
	s.Mutex.Unlock()
	return value
}

func (s *ServerConfig) GetParameters() map[string]Parameter {
	return s.parameters
}

func newServerConfig() *ServerConfig {
	return &ServerConfig{
		parameters: serverParams,
	}
}
