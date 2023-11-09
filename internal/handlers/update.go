package handlers

import (
	"net/http"
	"strconv"

	"github.com/ilya372317/must-have-metrics/internal/server/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
)

func UpdateHandler(storage storage.Storage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		updateAlertDTO := dto.CreateAlertDTOFromRequest(request)
		if err := addAlert(storage, updateAlertDTO); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}

func addAlert(repo storage.Storage, dto dto.UpdateAlertDTO) error {
	switch dto.Type {
	case entity.TypeGauge:
		err := updateGaugeAlert(dto, repo)
		if err != nil {
			return err
		}
	case entity.TypeCounter:
		err := updateCounterAlert(dto, repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateGaugeAlert(dto dto.UpdateAlertDTO, repository storage.Storage) error {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	repository.Save(dto.Name, alert)

	return nil
}

func updateCounterAlert(dto dto.UpdateAlertDTO, repo storage.Storage) error {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	if !repo.Has(dto.Name) {
		repo.Save(dto.Name, alert)
		return nil
	}
	oldAlert, err := repo.Get(dto.Name)
	if err != nil {
		return err
	}

	if oldAlert.Type == entity.TypeGauge {
		repo.Save(dto.Name, alert)
		return nil
	}

	newValue := oldAlert.Value.(int64) + alert.Value.(int64)
	alert.Value = newValue

	if err := repo.Update(dto.Name, alert); err != nil {
		return err
	}
	return nil
}
