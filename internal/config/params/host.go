package params

type HostConfig struct {
	Env string `env:"ADDRESS"`
	Parameter
}

func (h *HostConfig) GetEnv() string {
	return h.Env
}

func (h *HostConfig) SetEnv(s string) {
	h.Env = s
}

func (h *HostConfig) GetFlagName() string {
	return "a"
}

func (h *HostConfig) GetDefaultFlagValue() string {
	return "localhost:8080"
}

func (h *HostConfig) GetFlagDescription() string {
	return "server address"
}
