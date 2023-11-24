package params

type StoreIntervalConfig struct {
	Env string `env:"STORE_INTERVAL"`
	Parameter
}

func (s *StoreIntervalConfig) GetEnv() string {
	return s.Env
}

func (s *StoreIntervalConfig) SetEnv(str string) {
	s.Env = str
}

func (s *StoreIntervalConfig) GetFlagName() string {
	return "i"
}

func (s *StoreIntervalConfig) GetDefaultFlagValue() string {
	return "300"
}

func (s *StoreIntervalConfig) GetFlagDescription() string {
	return "interval saving metrics in file"
}
