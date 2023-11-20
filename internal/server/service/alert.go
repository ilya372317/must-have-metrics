package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const failedUpdateCounterPattern = "failed update counter alert: %w"
const failedParseGaugeValuePattern = "failed parse gauge alert value: %w"
const failedParseCounterValuePattern = "failed parse counter alert value: %w"

type UpdateStorage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
}

func AddAlert(repo UpdateStorage, dto dto.UpdateAlertDTO) (*entity.Alert, error) {
	switch dto.Type {
	case entity.TypeGauge:
		alert, err := updateGaugeAlert(dto, repo)
		if err != nil {
			return nil, fmt.Errorf("failed update gauge alert: %w", err)
		}
		return alert, nil
	case entity.TypeCounter:
		alert, err := updateCounterAlert(dto, repo)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return alert, nil

	default:
		return nil, errors.New("invalid type of metric")
	}
}

func updateGaugeAlert(dto dto.UpdateAlertDTO, repository UpdateStorage) (*entity.Alert, error) {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return nil, fmt.Errorf(failedParseGaugeValuePattern, err)
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	repository.Save(dto.Name, alert)
	newAlert, errAlertNotFound := repository.Get(dto.Name)
	if errAlertNotFound != nil {
		return nil, fmt.Errorf("failed update gauge alert: %w", err)
	}

	return &newAlert, nil
}

func updateCounterAlert(dto dto.UpdateAlertDTO, repo UpdateStorage) (*entity.Alert, error) {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(failedParseCounterValuePattern, err)
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	if !repo.Has(dto.Name) {
		repo.Save(dto.Name, alert)
		resultAlert, err := repo.Get(dto.Name)
		if err != nil {
			return nil, fmt.Errorf("failed update counter alert: %w", err)
		}
		return &resultAlert, nil
	}
	oldAlert, err := repo.Get(dto.Name)
	if err != nil {
		return nil, fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		repo.Save(dto.Name, alert)
		newAlert, err := repo.Get(dto.Name)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return &newAlert, nil
	}

	newValue := oldAlert.Value.(int64) + alert.Value.(int64)
	alert.Value = newValue

	if err := repo.Update(dto.Name, alert); err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := repo.Get(dto.Name)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return &newAlert, nil
}
