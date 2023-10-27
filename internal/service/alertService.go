package service

import (
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/repository"
)

func AddAlert(repo repository.AlertStorage, dto dto.UpdateAlertDTO) error {
	if err := repo.AddAlert(dto.Type, dto.Name, dto.Data); err != nil {
		return err
	}

	return nil
}
