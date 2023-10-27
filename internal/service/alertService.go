package service

import (
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/entity"
	"github.com/ilya372317/must-have-metrics/internal/repository"
	"strconv"
)

func AddAlert(repo repository.AlertStorage, dto dto.UpdateAlertDTO) error {
	switch dto.Type {
	case entity.GaugeType:
		err := updateGaugeAlert(dto, repo)
		if err != nil {
			return err
		}
	case entity.CounterType:
		err := updateCounterAlert(dto, repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateGaugeAlert(dto dto.UpdateAlertDTO, repo repository.AlertStorage) error {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	repo.SetAlert(dto.Name, alert)

	return nil
}

func updateCounterAlert(dto dto.UpdateAlertDTO, repo repository.AlertStorage) error {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return err
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	if !repo.HasAlert(dto.Name) {
		repo.SetAlert(dto.Name, alert)
		return nil
	}
	oldAlert, err := repo.GetAlert(dto.Name)
	if err != nil {
		return err
	}

	newValue := oldAlert.Value.Add(alert.Value)
	alert.Value = newValue

	if err := repo.UpdateAlert(dto.Name, alert); err != nil {
		return err
	}
	return nil
}
