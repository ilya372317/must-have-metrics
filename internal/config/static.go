package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type StaticConfig struct {
	RuleSet map[string]struct{}
}

func NewStaticConfig() *StaticConfig {
	viper.SetConfigName("staticchecks")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	cnfg := &StaticConfig{RuleSet: make(map[string]struct{})}
	untypedChecks, ok := viper.Get("checks").([]any)
	if !ok {
		panic(fmt.Errorf("invalid config content in checks field"))
	}

	checks := make([]string, 0, len(untypedChecks))
	for _, untypedCheck := range untypedChecks {
		value, ok := untypedCheck.(string)
		if !ok {
			panic("invalid config value in checks field")
		}
		checks = append(checks, value)
	}

	for _, check := range checks {
		cnfg.RuleSet[check] = struct{}{}
	}

	return cnfg
}
