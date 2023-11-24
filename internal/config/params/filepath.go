package params

type StoreFilePathConfig struct {
	Env string `env:"FILE_STORAGE_PATH"`
	Parameter
}

func (s *StoreFilePathConfig) GetEnv() string {
	return s.Env
}

func (s *StoreFilePathConfig) SetEnv(str string) {
	s.Env = str
}

func (s *StoreFilePathConfig) GetFlagName() string {
	return "f"
}

func (s *StoreFilePathConfig) GetDefaultFlagValue() string {
	return "/tmp/metrics-db.json"
}

func (s *StoreFilePathConfig) GetFlagDescription() string {
	return "file path where metrics will be stored"
}
