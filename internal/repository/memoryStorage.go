package repository

import (
	"github.com/ilya372317/must-have-metrics/internal/entity"
	"github.com/ilya372317/must-have-metrics/internal/errors"
)

type AlertInMemoryStorage struct {
	Records map[string]entity.Alert
}

func MakeAlertInMemoryStorage() AlertInMemoryStorage {
	return AlertInMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *AlertInMemoryStorage) AddAlert(name string, alert entity.Alert) {
	storage.Records[name] = alert
}

func (storage *AlertInMemoryStorage) UpdateAlert(name string, newValue entity.AlertValue) error {
	if !storage.HasAlert(name) {
		return &errors.AlertNotFound{}
	}
	currentAlert, err := storage.GetAlert(name)
	if err != nil {
		return err
	}
	currentAlert.Value = currentAlert.Value.Add(newValue)

	return nil
}

func (storage *AlertInMemoryStorage) GetAlert(name string) (entity.Alert, error) {
	alert, ok := storage.Records[name]
	if !ok {
		return entity.Alert{}, &errors.AlertNotFound{}
	}
	return alert, nil
}

func (storage *AlertInMemoryStorage) HasAlert(name string) bool {
	_, ok := storage.Records[name]
	return ok
}
