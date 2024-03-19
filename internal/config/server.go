package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

// ServerConfig server configs.
type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	SecretKey     string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	StoreInterval uint   `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
}

// NewServer constructor for ServerConfig.
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
		":8080", "address where server will listen requests",
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
	flag.StringVar(&c.SecretKey, "k", "", "Secret key for sign")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "Private crypto key for RSA decryption")
	flag.Parse()
}

// ShouldConnectToDatabase check for application configured to database connection.
func (c *ServerConfig) ShouldConnectToDatabase() bool {
	return c.DatabaseDSN != ""
}

// ShouldSignData check for server should sign response data.
func (c *ServerConfig) ShouldSignData() bool {
	return c.SecretKey != ""
}

// ShouldDecryptData check for server should decrypt request body.
func (c *ServerConfig) ShouldDecryptData() bool {
	if c.CryptoKey == "" {
		return false
	}

	if _, err := os.Stat(c.CryptoKey); os.IsNotExist(err) {
		return false
	}

	return true
}
