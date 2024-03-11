package logger

import (
	"fmt"
	"os"

	"github.com/ilya372317/must-have-metrics/internal/utils"
	"go.uber.org/zap"
)

// Log public logger instance.
var Log *zap.SugaredLogger

const (
	logPath           = "storage/log.txt"
	logFolder         = "storage"
	logFilePermission = 0750
)

// Init initialize logger for next using.
func Init() error {
	if Log != nil {
		return nil
	}
	createLogFolderIfNotExists()
	environment := os.Getenv("ENV")
	path := utils.Root + "/" + logPath
	if environment == "prod" {
		cnfg := zap.NewProductionConfig()
		cnfg.OutputPaths = []string{path, "stdout"}
		log, err := cnfg.Build()
		if err != nil {
			return fmt.Errorf("failed init zap logger in production: %w", err)
		}
		Log = log.Sugar()
	} else {
		cnfg := zap.NewDevelopmentConfig()
		cnfg.OutputPaths = []string{path, "stdout"}
		log, err := cnfg.Build()
		if err != nil {
			return fmt.Errorf("failed init zap logger in development: %w", err)
		}
		Log = log.Sugar()
	}

	return nil
}

func createLogFolderIfNotExists() {
	if _, err := os.Stat(utils.Root + "/" + logFolder); os.IsNotExist(err) {
		err = os.Mkdir(utils.Root+"/"+logFolder, logFilePermission)
		if err != nil {
			panic(fmt.Errorf("failed create log folder: %w", err))
		}
	}
}
