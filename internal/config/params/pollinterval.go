package params

type PollIntervalConfig struct {
	Env string `env:"POLL_INTERVAL"`
	Parameter
}

func (p *PollIntervalConfig) GetDefaultFlagValue() string {
	return "2"
}

func (p *PollIntervalConfig) GetFlagDescription() string {
	return "interval agent collect metrics"
}

func (p *PollIntervalConfig) GetFlagName() string {
	return "p"
}

func (p *PollIntervalConfig) GetEnv() string {
	return p.Env
}

func (p *PollIntervalConfig) SetEnv(s string) {
	p.Env = s
}
