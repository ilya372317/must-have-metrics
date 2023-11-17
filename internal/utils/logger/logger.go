package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

const (
	logPath           = "storage/log.txt"
	storageFolder     = "storage"
	storagePermission = 0750
)

func Get() *zap.SugaredLogger {
	if logger != nil {
		return logger
	}
	createStorageIfNotExists()
	environment := os.Getenv("ENV")
	if environment == "prod" {
		cnfg := zap.NewProductionConfig()
		cnfg.OutputPaths = []string{logPath}
		log, err := cnfg.Build()
		if err != nil {
			panic(fmt.Errorf("failed init zap logger in production: %w", err))
		}
		logger = log.Sugar()
	} else {
		cnfg := zap.NewDevelopmentConfig()
		cnfg.OutputPaths = []string{logPath}
		log, err := cnfg.Build()
		if err != nil {
			panic(fmt.Errorf("failed init zap logger in development: %w", err))
		}
		logger = log.Sugar()
	}

	return logger
}

func createStorageIfNotExists() {
	if _, err := os.Stat(storageFolder); os.IsNotExist(err) {
		err = os.Mkdir(storageFolder, storagePermission)
		if err != nil {
			panic(fmt.Errorf("failed create storage folder: %w", err))
		}
	}
}
