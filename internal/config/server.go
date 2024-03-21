package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	defaultServerHostValue          = ":8080"
	defaultServerFilePathValue      = "/tmp/metrics-db.json"
	defaultServerDatabaseDSNValue   = ""
	defaultServerRestoreValue       = true
	defaultServerStoreIntervalValue = 300
	defaultServerSecretKeyValue     = ""
	defaultServerCryptoKeyValue     = ""
	defaultServerConfigPathValue    = ""
)

// ServerConfig server configs.
type ServerConfig struct {
	Host          string `env:"ADDRESS" json:"address,omitempty"`
	FilePath      string `env:"FILE_STORAGE_PATH" json:"store_file,omitempty"`
	DatabaseDSN   string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	ConfigPath    string `env:"CONFIG"`
	SecretKey     string `env:"KEY" json:"secret_key,omitempty"`
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	StoreInterval uint   `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	Restore       bool   `env:"RESTORE" json:"restore,omitempty"`
}

// NewServer constructor for ServerConfig.
func NewServer() (*ServerConfig, error) {
	cnfg := &ServerConfig{}
	cnfg.parseFlags()
	err := env.Parse(cnfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse enviroment virables: %w", err)
	}
	if err = cnfg.parseFromFile(); err != nil {
		return nil, fmt.Errorf("failed create config from file: %w", err)
	}
	return cnfg, nil
}

func (c *ServerConfig) parseFlags() {
	flag.StringVar(
		&c.Host, "a",
		defaultServerHostValue, "address where server will listen requests",
	)
	flag.StringVar(
		&c.FilePath, "f",
		defaultServerFilePathValue, "file path where metrics will be stored",
	)
	flag.BoolVar(&c.Restore, "r", defaultServerRestoreValue, "Restore data from file in start server or not")
	flag.UintVar(&c.StoreInterval, "i", defaultServerStoreIntervalValue,
		"interval saving metrics in file",
	)
	flag.StringVar(&c.DatabaseDSN, "d", defaultServerDatabaseDSNValue, "Database DSN string")
	flag.StringVar(&c.SecretKey, "k", defaultServerSecretKeyValue, "Secret key for sign")
	flag.StringVar(&c.CryptoKey, "crypto-key", defaultServerCryptoKeyValue, "Private crypto key for RSA decryption")
	flag.StringVar(&c.ConfigPath, "c", defaultServerConfigPathValue, "file path to json configuration file")
	flag.Parse()
}

func (c *ServerConfig) parseFromFile() error {
	if c.ConfigPath == "" {
		return nil
	}

	configData, err := getConfigFileContent(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed get config content for server: %w", err)
	}

	tempConfig := ServerConfig{}

	if err = json.Unmarshal(configData, &tempConfig); err != nil {
		return fmt.Errorf("invalid data in server file config: %w", err)
	}

	if c.Host == defaultServerHostValue || c.Host == nullStringValue {
		c.Host = tempConfig.Host
	}

	if c.Restore {
		c.Restore = tempConfig.Restore
	}

	if c.StoreInterval == defaultServerStoreIntervalValue || c.StoreInterval == nullIntValue {
		c.StoreInterval = tempConfig.StoreInterval
	}

	if c.FilePath == defaultServerFilePathValue {
		c.FilePath = tempConfig.FilePath
	}

	if c.DatabaseDSN == defaultServerDatabaseDSNValue {
		c.DatabaseDSN = tempConfig.DatabaseDSN
	}

	if c.CryptoKey == defaultServerCryptoKeyValue {
		c.CryptoKey = tempConfig.CryptoKey
	}

	if c.SecretKey == defaultServerSecretKeyValue {
		c.SecretKey = tempConfig.SecretKey
	}

	return nil
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
