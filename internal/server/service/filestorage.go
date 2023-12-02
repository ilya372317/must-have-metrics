package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const filePermission = 0600

type FilesystemSupportStorage interface {
	AllWithKeys() map[string]entity.Alert
	Fill(map[string]entity.Alert)
}

func SaveDataToFilesystemByInterval(serverConfig *config.ServerConfig, repository FilesystemSupportStorage) {
	ticker := time.NewTicker(time.Duration(serverConfig.StoreInterval) * time.Second)
	for range ticker.C {
		err := StoreToFilesystem(repository, serverConfig.FilePath)
		if err != nil {
			logger.Get().Fatalf("failed save data to filesystem: %v", err)
			break
		}
	}
}

func FillFromFilesystem(storage FilesystemSupportStorage, filePath string) error {
	records := storage.AllWithKeys()
	if filePath == "" {
		return errors.New("no need to save data in filesystem")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed fill storage from file system: %w", err)
	}

	if err = json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("metrics in file is invalid: %w", err)
	}

	storage.Fill(records)

	return nil
}

func StoreToFilesystem(storage FilesystemSupportStorage, filepath string) error {
	data, err := json.Marshal(storage.AllWithKeys())
	if err != nil {
		return fmt.Errorf("failed serialize metrics: %w", err)
	}
	if err = os.WriteFile(filepath, data, filePermission); err != nil {
		return fmt.Errorf("failed save file on disk: %w", err)
	}

	return nil
}
