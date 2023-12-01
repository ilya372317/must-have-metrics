package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
	"github.com/joho/godotenv"
)

var sLogger = logger.Get()

func main() {
	if err := run(); err != nil {
		sLogger.Panicf("failed to start server: %v", err)
	}
}

func run() error {
	repository := storage.NewInMemoryStorage()
	if err := godotenv.Load(".env-server"); err != nil {
		logger.Get().Warnf("failed load .env-server file: %v", err)
	}
	cnfg, err := config.NewServer()
	if err != nil {
		sLogger.Panicf("failed get server config: %v", err)
	}
	if cnfg.StoreInterval > 0 {
		go service.SaveDataToFilesystemByInterval(
			time.Duration(cnfg.StoreInterval)*time.Second,
			cnfg.FilePath,
			repository,
		)
	}
	if cnfg.Restore {
		if err = repository.FillFromFilesystem(cnfg.FilePath); err != nil {
			sLogger.Warn(err)
		}
	}
	sLogger.Infof("server is starting...")
	err = http.ListenAndServe(
		cnfg.Host,
		router.AlertRouter(repository, cnfg),
	)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
