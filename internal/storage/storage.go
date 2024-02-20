package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

var _ handlers.IndexStorage = (*InMemoryStorage)(nil)

type AlertNotFoundError struct{}

func (e *AlertNotFoundError) Error() string {
	return "alert not found"
}

type InMemoryStorage struct {
	Records map[string]entity.Alert `json:"records"`
	sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *InMemoryStorage) Save(_ context.Context, name string, alert entity.Alert) error {
	storage.Records[name] = alert
	return nil
}

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

func (storage *InMemoryStorage) Get(_ context.Context, name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &AlertNotFoundError{}
	}
	return alert, nil
}

func (storage *InMemoryStorage) Has(_ context.Context, name string) (bool, error) {
	_, ok := storage.Records[name]
	return ok, nil
}

func (storage *InMemoryStorage) All(context.Context) ([]entity.Alert, error) {
	values := make([]entity.Alert, 0, len(storage.Records))
	for _, value := range storage.Records {
		values = append(values, value)
	}

	return values, nil
}

func (storage *InMemoryStorage) AllWithKeys(context.Context) (map[string]entity.Alert, error) {
	return storage.Records, nil
}

func (storage *InMemoryStorage) Fill(_ context.Context, alerts map[string]entity.Alert) error {
	storage.Records = alerts
	return nil
}

func (storage *InMemoryStorage) Reset() {
	storage.Records = make(map[string]entity.Alert)
}

func (storage *InMemoryStorage) BulkInsertOrUpdate(_ context.Context, alerts []entity.Alert) error {
	storage.Mutex.Lock()
	for _, alert := range alerts {
		storage.Records[alert.Name] = alert
	}

	storage.Mutex.Unlock()

	return nil
}

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

func (storage *InMemoryStorage) Ping() error {
	return nil
}
