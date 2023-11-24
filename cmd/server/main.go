package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var (
	repository *storage.InMemoryStorage
	sLogger    = logger.Get()
)

func main() {
	if err := config.InitServerConfig(); err != nil {
		sLogger.Panicf("failed start server: %v", err)
	}
	repository = storage.NewInMemoryStorage()

	cnfg := config.GetServerConfig()
	isRestart, err := strconv.ParseBool(cnfg.GetValue("restore"))
	if err != nil {
		sLogger.Panicf("invalid restart configuration value: %v", err)
	}
	if isRestart {
		if err = repository.FillFromFilesystem(); err != nil {
			sLogger.Warn(err)
		}
	}

	storeInterval, err := strconv.Atoi(cnfg.GetValue("store_interval"))
	if err != nil {
		sLogger.Panicf("failed parse store interval parameter")
	}

	if storeInterval > 0 {
		go service.SaveDataToFilesystemByInterval(
			time.Duration(storeInterval)*time.Second,
			cnfg.GetValue("store_path"),
			repository,
		)
	}

	if err = run(cnfg, storeInterval == 0, cnfg.GetValue("store_path")); err != nil {
		sLogger.Panicf("failed to start server on address: %s, %v", cnfg.GetValue("host"), err)
	}
}

func run(cnfg *config.ServerConfig, isSyncSaving bool, fileStoragePath string) error {
	sLogger.Infof("server is starting...")
	err := http.ListenAndServe(cnfg.GetValue("host"), router.AlertRouter(repository, isSyncSaving, fileStoragePath))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
