package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type ErrAlertNotFound struct{}

func (e *ErrAlertNotFound) Error() string {
	return "alert not found"
}

type InMemoryStorage struct {
	sync.Mutex
	Records map[string]entity.Alert `json:"records"`
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *InMemoryStorage) Save(name string, alert entity.Alert) {
	storage.Records[name] = alert
}

func (storage *InMemoryStorage) Update(name string, newValue entity.Alert) error {
	if !storage.Has(name) {
		return &ErrAlertNotFound{}
	}
	storage.Save(name, newValue)

	return nil
}

func (storage *InMemoryStorage) Get(name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &ErrAlertNotFound{}
	}
	return alert, nil
}

func (storage *InMemoryStorage) Has(name string) bool {
	_, ok := storage.Records[name]
	return ok
}

func (storage *InMemoryStorage) All() []entity.Alert {
	values := make([]entity.Alert, 0, len(storage.Records))
	for _, value := range storage.Records {
		values = append(values, value)
	}

	return values
}

func (storage *InMemoryStorage) Reset() {
	storage.Records = make(map[string]entity.Alert)
}

func (storage *InMemoryStorage) FillFromFilesystem(filePath string) error {
	if filePath == "" {
		return errors.New("no need to save data in filesystem")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed fill storage from file system: %w", err)
	}

	if err = json.Unmarshal(data, storage); err != nil {
		return fmt.Errorf("metrics in file is invalid: %w", err)
	}

	return nil
}

func (storage *InMemoryStorage) StoreToFilesystem(filepath string) error {
	storage.Lock()
	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed serialize metrics: %w", err)
	}
	if err = os.WriteFile(filepath, data, 0666); err != nil {
		return fmt.Errorf("failed save file on disk: %w", err)
	}
	storage.Unlock()

	return nil
}
