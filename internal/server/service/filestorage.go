package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const filePermission = 0600

type filesystemSupportStorage interface {
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
}

// SaveDataToFilesystemByInterval by configured interval saving data from storage to filesystem.
func SaveDataToFilesystemByInterval(
	ctx context.Context,
	wg *sync.WaitGroup,
	serverConfig *config.ServerConfig, repository filesystemSupportStorage) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(serverConfig.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := StoreToFilesystem(ctx, repository, serverConfig.FilePath)
			if err != nil {
				logger.Log.Errorf("failed save data to filesystem: %v", err)
				return
			}
		}
	}
}

// FillFromFilesystem truncate storage and save data from filesystem to storage.
func FillFromFilesystem(ctx context.Context, storage filesystemSupportStorage, filePath string) error {
	records, err := storage.AllWithKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed get all records with keys: %w", err)
	}
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

	if err = storage.Fill(ctx, records); err != nil {
		return fmt.Errorf("failed fill items: %w", err)
	}

	return nil
}

// StoreToFilesystem get all records from storage and save them to filesystem.
func StoreToFilesystem(ctx context.Context, storage filesystemSupportStorage, filepath string) error {
	allItemsWithKeys, err := storage.AllWithKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed get all items: %w", err)
	}
	data, err := json.Marshal(allItemsWithKeys)
	if err != nil {
		return fmt.Errorf("failed serialize metrics: %w", err)
	}
	if err = os.WriteFile(filepath, data, filePermission); err != nil {
		return fmt.Errorf("failed save file on disk: %w", err)
	}

	return nil
}
