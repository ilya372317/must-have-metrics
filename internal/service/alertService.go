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
		floatData, err := strconv.ParseFloat(dto.Data, 64)
		if err != nil {
			return err
		}
		alert := entity.MakeGaugeAlert(dto.Name, floatData)
		repo.AddAlert(dto.Name, alert)
		break
	case entity.CounterType:
		intData, err := strconv.ParseInt(dto.Data, 10, 64)
		if err != nil {
			return err
		}
		alert := entity.MakeCounterAlert(dto.Name, intData)
		if err := updateCounterAlert(dto.Name, repo, alert); err != nil {
			return err
		}
		break
	}

	return nil
}

func updateCounterAlert(name string, repo repository.AlertStorage, alert entity.Alert) error {
	if !repo.HasAlert(name) {
		repo.AddAlert(name, alert)
		return nil
	}
	if err := repo.UpdateAlert(name, alert.Value); err != nil {
		return err
	}
	return nil
}
