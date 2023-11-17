package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"

	"github.com/ilya372317/must-have-metrics/internal/server/dto"
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

func UpdateHandler(storage UpdateStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		updateAlertDTO := dto.CreateUpdateAlertDTOFromRequest(request)
		_, err := updateAlertDTO.Validate()
		if err != nil {
			http.Error(writer, fmt.Errorf("invalid parameters: %w", err).Error(), http.StatusBadRequest)
		}
		if err := addAlert(storage, updateAlertDTO); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}

func addAlert(repo UpdateStorage, dto dto.UpdateAlertDTO) error {
	switch dto.Type {
	case entity.TypeGauge:
		err := updateGaugeAlert(dto, repo)
		if err != nil {
			return fmt.Errorf("failed update gauge alert: %w", err)
		}
	case entity.TypeCounter:
		err := updateCounterAlert(dto, repo)
		if err != nil {
			return fmt.Errorf(failedUpdateCounterPattern, err)
		}
	}

	return nil
}

func updateGaugeAlert(dto dto.UpdateAlertDTO, repository UpdateStorage) error {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return fmt.Errorf(failedParseGaugeValuePattern, err)
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	repository.Save(dto.Name, alert)

	return nil
}

func updateCounterAlert(dto dto.UpdateAlertDTO, repo UpdateStorage) error {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return fmt.Errorf(failedParseCounterValuePattern, err)
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	if !repo.Has(dto.Name) {
		repo.Save(dto.Name, alert)
		return nil
	}
	oldAlert, err := repo.Get(dto.Name)
	if err != nil {
		return fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		repo.Save(dto.Name, alert)
		return nil
	}

	newValue := oldAlert.Value.(int64) + alert.Value.(int64)
	alert.Value = newValue

	if err := repo.Update(dto.Name, alert); err != nil {
		return fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return nil
}
