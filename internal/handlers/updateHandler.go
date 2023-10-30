package handlers

import (
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"github.com/ilya372317/must-have-metrics/internal/server/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"net/http"
	"strconv"
)

func UpdateHandler(storage storage.AlertStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		updateAlertDTO := dto.CreateAlertDTOFromRequest(request)
		if err := addAlert(storage, updateAlertDTO); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}

func addAlert(repo storage.AlertStorage, dto dto.UpdateAlertDTO) error {
	switch dto.Type {
	case constant.TypeGauge:
		err := updateGaugeAlert(dto, repo)
		if err != nil {
			return err
		}
	case constant.TypeCounter:
		err := updateCounterAlert(dto, repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateGaugeAlert(dto dto.UpdateAlertDTO, repository storage.AlertStorage) error {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	repository.SaveAlert(dto.Name, alert)

	return nil
}

func updateCounterAlert(dto dto.UpdateAlertDTO, repo storage.AlertStorage) error {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	if !repo.HasAlert(dto.Name) {
		repo.SaveAlert(dto.Name, alert)
		return nil
	}
	oldAlert, err := repo.GetAlert(dto.Name)
	if err != nil {
		return err
	}

	newValue := oldAlert.Value.(int64) + alert.Value.(int64)
	alert.Value = newValue

	if err := repo.UpdateAlert(dto.Name, alert); err != nil {
		return err
	}
	return nil
}
