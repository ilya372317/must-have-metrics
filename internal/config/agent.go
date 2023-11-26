package config

import (
	"fmt"
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var agentConfig *AgentConfig

type AgentConfig struct {
	sync.Mutex
	parameters map[string]Parameter
}

func GetAgentConfig() *AgentConfig {
	if agentConfig == nil {
		logger.Get().Panicf("You forget init agent configuration! Please, do it in cmd/agent/main.go")
	}

	return agentConfig
}

func InitAgentConfig() error {
	if agentConfig != nil {
		return nil
	}

	agentConfig = newAgentConfig()
	if err := initConfiguration(agentConfig, false); err != nil {
		serverConfig = nil
		return fmt.Errorf("failed init agent configuration: %w", err)
	}

	return nil
}

func (a *AgentConfig) GetValue(alias string) string {
	value := ""
	a.Mutex.Lock()
	value = a.parameters[alias].GetValue()
	a.Mutex.Unlock()
	return value
}

func (a *AgentConfig) GetParameters() map[string]Parameter {
	return a.parameters
}

func newAgentConfig() *AgentConfig {
	return &AgentConfig{
		parameters: agentParams,
	}
}
