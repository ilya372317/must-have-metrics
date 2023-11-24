package params

type RestoreConfig struct {
	Env string `env:"RESTORE"`
	Parameter
}

func (r *RestoreConfig) GetEnv() string {
	return r.Env
}

func (r *RestoreConfig) SetEnv(s string) {
	r.Env = s
}

func (r *RestoreConfig) GetFlagName() string {
	return "r"
}

func (r *RestoreConfig) GetDefaultFlagValue() string {
	return "true"
}

func (r *RestoreConfig) GetFlagDescription() string {
	return "Restore data from file in start server or not"
}
