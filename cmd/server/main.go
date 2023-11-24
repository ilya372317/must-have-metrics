package main

import (
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var (
	repository *storage.InMemoryStorage
	sLogger    = logger.Get()
)

func main() {
	cnfg := config.NewServerConfig()
	repository = storage.NewInMemoryStorage()
	if err := cnfg.Init(); err != nil {
		sLogger.Panicf("failed parse config: %v", err)
	}

	if err := run(cnfg); err != nil {
		sLogger.Panicf("failed to start server on address: %s, %v", cnfg.GetValue("host"), err)
	}
}

func run(cnfg *config.ServerConfig) error {
	sLogger.Infof("server is starting...")
	err := http.ListenAndServe(cnfg.GetValue("host"), router.AlertRouter(repository))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
