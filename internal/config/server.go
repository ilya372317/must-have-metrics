package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	StoreInterval uint   `env:"STORE_INTERVAL"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
}

func NewServer() (*ServerConfig, error) {
	cnfg := &ServerConfig{}
	cnfg.parseFlags()
	err := env.Parse(cnfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse enviroment virables: %w", err)
	}
	return cnfg, nil
}

func (c *ServerConfig) parseFlags() {
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
	flag.StringVar(&c.DatabaseDSN, "d", "", "Database DSN string")
	flag.Parse()
}
