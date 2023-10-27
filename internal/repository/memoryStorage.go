package repository

import (
	"github.com/ilya372317/must-have-metrics/internal/entity"
	"github.com/ilya372317/must-have-metrics/internal/errors"
)

type InMemoryStorage struct {
	Records map[string]entity.Alert
}

func MakeAlertInMemoryStorage() InMemoryStorage {
	return InMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *InMemoryStorage) SetAlert(name string, alert entity.Alert) {
	storage.Records[name] = alert
}

func (storage *InMemoryStorage) UpdateAlert(name string, newValue entity.Alert) error {
	if !storage.HasAlert(name) {
		return &errors.AlertNotFound{}
	}
	storage.Records[name] = newValue

	return nil
}

func (storage *InMemoryStorage) GetAlert(name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &errors.AlertNotFound{}
	}
	return alert, nil
}

func (storage *InMemoryStorage) HasAlert(name string) bool {
	_, ok := storage.Records[name]
	return ok
}
