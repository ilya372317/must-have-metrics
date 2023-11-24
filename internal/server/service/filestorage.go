package service

import (
	"time"

	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

type FilesystemSupportStorage interface {
	StoreToFilesystem(filepath string) error
}

func SaveDataToFilesystemByInterval(interval time.Duration, filepath string, repository FilesystemSupportStorage) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		err := repository.StoreToFilesystem(filepath)
		if err != nil {
			logger.Get().Fatalf("failed save data to filesystem: %v", err)
			break
		}
	}
}
