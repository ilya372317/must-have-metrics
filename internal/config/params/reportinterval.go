package params

type ReportIntervalConfig struct {
	Env string `env:"REPORT_INTERVAL"`
	Parameter
}

func (r *ReportIntervalConfig) GetFlagName() string {
	return "r"
}

func (r *ReportIntervalConfig) GetDefaultFlagValue() string {
	return "10"
}

func (r *ReportIntervalConfig) GetFlagDescription() string {
	return "interval agent send metrics on server"
}

func (r *ReportIntervalConfig) GetEnv() string {
	return r.Env
}

func (r *ReportIntervalConfig) SetEnv(s string) {
	r.Env = s
}
