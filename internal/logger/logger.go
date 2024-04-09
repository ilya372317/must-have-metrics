package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
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

func InterceptorLogger() logging.Logger {
	if Log == nil {
		return nil
	}
	l := Log
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]interface{}, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
