package repository

import (
	"github.com/ilya372317/must-have-metrics/internal/entity"
	"strconv"
)

type AlertInMemoryStorage struct {
	Records map[string]entity.Alert
}

func MakeAlertInMemoryStorage() AlertInMemoryStorage {
	return AlertInMemoryStorage{
		Records: make(map[string]entity.Alert),
	}
}

func (storage *AlertInMemoryStorage) AddAlert(typ, name, data string) error {
	switch typ {
	case entity.GaugeType:
		floatData, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return err
		}
		storage.addGaugeAlert(typ, name, floatData)
		break
	case entity.CounterType:
		intData, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}
		storage.addCounterAlert(typ, name, intData)
		break
	}

	return nil
}

func (storage *AlertInMemoryStorage) addGaugeAlert(typ, name string, data float64) {
	storage.Records[name] = entity.MakeGaugeAlert(typ, name, data)
}

func (storage *AlertInMemoryStorage) addCounterAlert(typ, name string, data int64) {
	if _, ok := storage.Records[name]; !ok {
		storage.Records[name] = entity.MakeCounterAlert(typ, name, data)
		return
	}
	alertValue := storage.Records[name].Value
	storage.Records[name].Value.SetIntValue(alertValue.GetIntValue() + data)
}

func (storage *AlertInMemoryStorage) GetAlert(name string) entity.Alert {
	return entity.Alert{}
}
