package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const failedUpdateCounterPattern = "failed update counter alert: %w"
const failedParseGaugeValuePattern = "failed parse gauge alert value: %w"
const failedParseCounterValuePattern = "failed parse counter alert value: %w"
const failedUpdateGaugeAlertPattern = "failed update gauge alert: %w"

type UpdateStorage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
	AllWithKeys() map[string]entity.Alert
	Fill(map[string]entity.Alert)
}

func AddAlert(
	repo UpdateStorage,
	dto dto.UpdateAlertDTO,
	serverConfig *config.ServerConfig,
) (*entity.Alert, error) {
	var alert *entity.Alert
	var err error
	switch dto.Type {
	case entity.TypeGauge:
		alert, err = updateGaugeAlert(dto, repo)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
		break
	case entity.TypeCounter:
		alert, err = updateCounterAlert(dto, repo)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		break
	default:
		return nil, errors.New("invalid type of metric")
	}

	if err = StoreToFilesystem(repo, serverConfig.FilePath); err != nil {
		return nil, fmt.Errorf("failed save data to filesystem: %w", err)
	}

	return alert, nil
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
		return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
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

	newValue := *oldAlert.IntValue + *alert.IntValue
	alert.IntValue = &newValue

	if err := repo.Update(dto.Name, alert); err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := repo.Get(dto.Name)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return &newAlert, nil
}
