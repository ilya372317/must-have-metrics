package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
	"github.com/joho/godotenv"
)

const defaultServerAddress = "localhost:8080"
const staticFilePath = "static"

var (
	repository *storage.InMemoryStorage
	host       *string
	sLogger    = logger.Get()
)

func init() {
	repository = storage.NewInMemoryStorage()
	cnfg := new(config.ServerConfig)
	if err := cnfg.Init(); err != nil {
		log.Fatalln(err.Error())
	}
	host = flag.String("a", defaultServerAddress, "server address")

	if cnfg.Host != "" {
		host = &cnfg.Host
	}
	if err := godotenv.Load(".env-server"); err != nil {
		log.Panic(fmt.Errorf("failed to load env file: %w", err))
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		sLogger.Fatalf("failed to start server on address: %s, %v", *host, err)
	}
}

func run() error {
	sLogger.Infof("server is starting...")
	err := http.ListenAndServe(*host, router.AlertRouter(repository, staticFilePath))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
