package config

import (
	"fmt"
	"sync"
)

var agentConfig *AgentConfig

type AgentConfig struct {
	parameters map[string]Parameter
	sync.Mutex
}

func GetAgentConfig() (*AgentConfig, error) {
	if agentConfig != nil {
		return agentConfig, nil
	}
	agentConfig = newAgentConfig()
	err := initConfiguration(agentConfig, false)
	if err != nil {
		agentConfig = nil
		return nil, fmt.Errorf("failed get agent configuration: %w", err)
	}

	return agentConfig, nil
}

func (a *AgentConfig) GetValue(alias string) string {
	a.Mutex.Lock()
	value := a.parameters[alias].GetValue()
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
