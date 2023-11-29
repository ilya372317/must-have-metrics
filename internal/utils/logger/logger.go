package logger

import (
	"fmt"
	"os"

	"github.com/ilya372317/must-have-metrics/internal/utils"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

const (
	logPath           = "storage/log.txt"
	logFolder         = "storage"
	logFilePermission = 0750
)

func Get() *zap.SugaredLogger {
	if logger != nil {
		return logger
	}
	createLogFolderIfNotExists()
	environment := os.Getenv("ENV")
	path := utils.BasePath() + "/" + logPath
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
	if _, err := os.Stat(utils.BasePath() + "/" + logFolder); os.IsNotExist(err) {
		err = os.Mkdir(utils.BasePath()+"/"+logFolder, logFilePermission)
		if err != nil {
			panic(fmt.Errorf("failed create log folder: %w", err))
		}
	}
}
