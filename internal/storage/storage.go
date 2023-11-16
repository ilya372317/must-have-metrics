package storage

import (
	"errors"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

var errAlertNotFound = errors.New("alert not found")

type InMemoryStorage struct {
	records map[string]entity.Alert
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		records: make(map[string]entity.Alert),
	}
}

func (storage *InMemoryStorage) Save(name string, alert entity.Alert) {
	storage.records[name] = alert
}

func (storage *InMemoryStorage) Update(name string, newValue entity.Alert) error {
	if !storage.Has(name) {
		return errAlertNotFound
	}
	storage.Save(name, newValue)

	return nil
}

func (storage *InMemoryStorage) Get(name string) (entity.Alert, error) {
	alert, ok := storage.records[name]
	if !ok {
		return entity.Alert{}, errAlertNotFound
	}
	return alert, nil
}

func (storage *InMemoryStorage) Has(name string) bool {
	_, ok := storage.records[name]
	return ok
}

func (storage *InMemoryStorage) All() []entity.Alert {
	values := make([]entity.Alert, 0, len(storage.records))
	for _, value := range storage.records {
		values = append(values, value)
	}

	return values
}
