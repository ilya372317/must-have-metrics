package config

import (
	"fmt"
	"sync"
)

var serverConfig *ServerConfig

type ServerConfig struct { //nolint:govet // how fix it i don`t now
	sync.Mutex
	parameters map[string]Parameter
}

func GetServerConfig() (*ServerConfig, error) {
	if serverConfig != nil {
		return serverConfig, nil
	}
	serverConfig = newServerConfig()
	if err := initConfiguration(serverConfig, true); err != nil {
		serverConfig = nil
		return nil, fmt.Errorf("failed init agent config: %w", err)
	}

	return serverConfig, nil
}

func (s *ServerConfig) GetValue(alias string) string {
	s.Mutex.Lock()
	value := s.parameters[alias].GetValue()
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
