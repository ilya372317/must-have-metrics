package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

const (
	logPath           = "storage/log.txt"
	logFolder         = "storage"
	logFilePermission = 0750
	basePath          = "../.."
)

func Get() *zap.SugaredLogger {
	if logger != nil {
		return logger
	}
	createLogFolderIfNotExists()
	environment := os.Getenv("ENV")
	path := basePath + "/" + logPath
	if environment == "prod" {
		cnfg := zap.NewProductionConfig()
		cnfg.OutputPaths = []string{path, "stdout"}
		log, err := cnfg.Build()
		if err != nil {
			panic(fmt.Errorf("failed init zap logger in production: %w", err))
		}
		logger = log.Sugar()
	} else {
		cnfg := zap.NewDevelopmentConfig()
		cnfg.OutputPaths = []string{path, "stdout"}
		log, err := cnfg.Build()
		if err != nil {
			panic(fmt.Errorf("failed init zap logger in development: %w", err))
		}
		logger = log.Sugar()
	}

	return logger
}

func createLogFolderIfNotExists() {
	if _, err := os.Stat(basePath + "/" + logFolder); os.IsNotExist(err) {
		err = os.Mkdir(basePath+"/"+logFolder, logFilePermission)
		if err != nil {
			panic(fmt.Errorf("failed create log folder: %w", err))
		}
	}
}
