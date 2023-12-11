package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const failedUpdateCounterPattern = "failed update counter alert: %w"
const failedUpdateGaugeAlertPattern = "failed update gauge alert: %w"

type UpdateStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
}

func AddAlert(
	ctx context.Context,
	repo UpdateStorage,
	dto dto.Metrics,
	serverConfig *config.ServerConfig,
) (*entity.Alert, error) {
	var alert *entity.Alert
	var err error
	switch dto.MType {
	case entity.TypeGauge:
		alert, err = updateGaugeAlert(ctx, dto, repo)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	case entity.TypeCounter:
		alert, err = updateCounterAlert(ctx, dto, repo)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
	default:
		return nil, errors.New("invalid type of metric")
	}

	if serverConfig.StoreInterval == 0 {
		if err = StoreToFilesystem(ctx, repo, serverConfig.FilePath); err != nil {
			return nil, fmt.Errorf("failed save data to filesystem: %w", err)
		}
	}

	return alert, nil
}

func updateGaugeAlert(ctx context.Context, dto dto.Metrics, repository UpdateStorage) (*entity.Alert, error) {
	if dto.Value == nil {
		return nil, errors.New("failed update gauge alert, missing value field")
	}

	alert := entity.MakeGaugeAlert(dto.ID, *dto.Value)
	alertExist, err := repository.Has(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
	}
	if alertExist {
		if err = repository.Update(ctx, dto.ID, alert); err != nil {
			return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	} else {
		if err = repository.Save(ctx, dto.ID, alert); err != nil {
			return nil, fmt.Errorf("failed save alert: %w", err)
		}
	}

	newAlert, errAlertNotFound := repository.Get(ctx, dto.ID)
	if errAlertNotFound != nil {
		return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, errAlertNotFound)
	}

	return &newAlert, nil
}

func updateCounterAlert(ctx context.Context, dto dto.Metrics, repo UpdateStorage) (*entity.Alert, error) {
	if dto.Delta == nil {
		return nil, errors.New("failed update counter alert, missing delta filed")
	}

	alert := entity.MakeCounterAlert(dto.ID, *dto.Delta)
	hasAlert, err := repo.Has(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf("failed check alert exist: %w", err)
	}
	if !hasAlert {
		if err = repo.Save(ctx, dto.ID, alert); err != nil {
			return nil, fmt.Errorf("failed save gauge alert: %w", err)
		}

		resultAlert, err := repo.Get(ctx, dto.ID)
		if err != nil {
			return nil, fmt.Errorf("failed update counter alert: %w", err)
		}
		return &resultAlert, nil
	}
	oldAlert, err := repo.Get(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		if err = repo.Save(ctx, dto.ID, alert); err != nil {
			return nil, fmt.Errorf("failed save gauge alert: %w", err)
		}
		newAlert, err := repo.Get(ctx, dto.ID)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return &newAlert, nil
	}

	newValue := *oldAlert.IntValue + *alert.IntValue
	alert.IntValue = &newValue

	if err := repo.Update(ctx, dto.ID, alert); err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := repo.Get(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return &newAlert, nil
}
