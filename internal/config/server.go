package config

import "sync"

type ServerConfig struct {
	sync.Mutex
	parameters map[string]Parameter
}

func (s *ServerConfig) GetValue(alias string) string {
	value := ""
	s.Mutex.Lock()
	value = s.parameters[alias].GetValue()
	s.Mutex.Unlock()
	return value
}

func (s *ServerConfig) Init() error {
	return initConfiguration(s)
}

func (s *ServerConfig) GetParameters() map[string]Parameter {
	return s.parameters
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		parameters: serverParams,
	}
}
