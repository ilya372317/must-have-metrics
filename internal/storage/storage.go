package storage

import (
	"errors"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

var alertNotFound = errors.New("alert not found")

type Storage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
	All() []entity.Alert
}

type InMemoryStorage struct {
	Records map[string]entity.Alert
}

func MakeInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *InMemoryStorage) Save(name string, alert entity.Alert) {
	storage.Records[name] = alert
}

func (storage *InMemoryStorage) Update(name string, newValue entity.Alert) error {
	if !storage.Has(name) {
		return alertNotFound
	}
	storage.Save(name, newValue)

	return nil
}

func (storage *InMemoryStorage) Get(name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, alertNotFound
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
