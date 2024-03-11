package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type AlertNotFoundError struct{}

func (e *AlertNotFoundError) Error() string {
	return "alert not found"
}

// InMemoryStorage storage representing data in memory.
type InMemoryStorage struct {
	Records map[string]entity.Alert `json:"records"`
	sync.Mutex
}

// NewInMemoryStorage constructor for InMemoryStorage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Records: make(map[string]entity.Alert, 1000),
	}
}

// Save saving record to memory.
func (storage *InMemoryStorage) Save(_ context.Context, name string, alert entity.Alert) error {
	storage.Records[name] = alert
	return nil
}

// Update updating record in memory.
func (storage *InMemoryStorage) Update(_ context.Context, name string, newValue entity.Alert) error {
	storageHasRecord, err := storage.Has(context.Background(), name)
	if err != nil {
		return fmt.Errorf("forbidden change existing value: %w", err)
	}

	if !storageHasRecord {
		return &AlertNotFoundError{}
	}
	if err = storage.Save(context.Background(), name, newValue); err != nil {
		return fmt.Errorf("failed save new value: %w", err)
	}

	return nil
}

// Get getting record from memory.
func (storage *InMemoryStorage) Get(_ context.Context, name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &AlertNotFoundError{}
	}
	return alert, nil
}

// Has checked if record with given name existed in memory.
func (storage *InMemoryStorage) Has(_ context.Context, name string) (bool, error) {
	_, ok := storage.Records[name]
	return ok, nil
}

// All retrieve all records from memory.
func (storage *InMemoryStorage) All(context.Context) ([]entity.Alert, error) {
	values := make([]entity.Alert, 0, len(storage.Records))
	for _, value := range storage.Records {
		values = append(values, value)
	}

	return values, nil
}

// AllWithKeys retrieve all records from memory in map representation.
func (storage *InMemoryStorage) AllWithKeys(context.Context) (map[string]entity.Alert, error) {
	return storage.Records, nil
}

// Fill delete all records from memory and saving given.
func (storage *InMemoryStorage) Fill(_ context.Context, alerts map[string]entity.Alert) error {
	storage.Records = alerts
	return nil
}

// Reset delete all data from memory.
func (storage *InMemoryStorage) Reset() {
	storage.Records = make(map[string]entity.Alert)
}

// BulkInsertOrUpdate if record representing in memory updating it. Otherwise, save it.
func (storage *InMemoryStorage) BulkInsertOrUpdate(_ context.Context, alerts []entity.Alert) error {
	storage.Mutex.Lock()
	for _, alert := range alerts {
		storage.Records[alert.Name] = alert
	}

	storage.Mutex.Unlock()

	return nil
}

// GetByIDs retrieve all records from memory by given ids.
func (storage *InMemoryStorage) GetByIDs(_ context.Context, ids []string) ([]entity.Alert, error) {
	storage.Mutex.Lock()
	resultAlerts := make([]entity.Alert, 0, len(ids))

	for _, id := range ids {
		alert, ok := storage.Records[id]
		if ok {
			resultAlerts = append(resultAlerts, alert)
		}
	}

	storage.Mutex.Unlock()

	return resultAlerts, nil
}

// Ping check if connection with storage is ok. In this case, it always ok.
func (storage *InMemoryStorage) Ping() error {
	return nil
}
