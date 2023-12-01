package storage

import (
	"sync"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

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

func (storage *InMemoryStorage) Save(name string, alert entity.Alert) {
	storage.Records[name] = alert
}

func (storage *InMemoryStorage) Update(name string, newValue entity.Alert) error {
	if !storage.Has(name) {
		return &AlertNotFoundError{}
	}
	storage.Save(name, newValue)

	return nil
}

func (storage *InMemoryStorage) Get(name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &AlertNotFoundError{}
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

func (storage *InMemoryStorage) AllWithKeys() map[string]entity.Alert {
	return storage.Records
}

func (storage *InMemoryStorage) Fill(alerts map[string]entity.Alert) {
	storage.Records = alerts
}

func (storage *InMemoryStorage) Reset() {
	storage.Records = make(map[string]entity.Alert)
}
