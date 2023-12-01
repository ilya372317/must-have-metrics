package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Host          string `env:"ADDRESS"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	StoreInterval uint   `env:"STORE_INTERVAL"`
}

func NewServer() (*Config, error) {
	cnfg := &Config{}
	cnfg.parseFlags()
	err := env.Parse(cnfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse enviroment virables: %w", err)
	}
	return cnfg, nil
}

func (c *Config) parseFlags() {
	flag.StringVar(
		&c.Host, "a",
		"localhost:8080", "address where server will listen requests",
	)
	flag.StringVar(
		&c.FilePath, "f",
		"/tmp/metrics-db.json", "file path where metrics will be stored",
	)
	flag.BoolVar(&c.Restore, "r", true, "Restore data from file in start server or not")
	flag.UintVar(&c.StoreInterval, "i", 300,
		"interval saving metrics in file",
	)
	flag.Parse()
}
