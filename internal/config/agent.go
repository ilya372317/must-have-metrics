package config

import "sync"

type AgentConfig struct {
	sync.Mutex
	parameters map[string]Parameter
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

func (a *AgentConfig) Init() error {
	return initConfiguration(a)
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		parameters: agentParams,
	}
}
