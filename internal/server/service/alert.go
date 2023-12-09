package service

import (
	"context"
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
	dto dto.UpdateAlertDTO,
	serverConfig *config.ServerConfig,
) (*entity.Alert, error) {
	var alert *entity.Alert
	var err error
	switch dto.Type {
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

func updateGaugeAlert(ctx context.Context, dto dto.UpdateAlertDTO, repository UpdateStorage) (*entity.Alert, error) {
	floatData, err := strconv.ParseFloat(dto.Data, 64)
	if err != nil {
		return nil, fmt.Errorf(failedParseGaugeValuePattern, err)
	}
	alert := entity.MakeGaugeAlert(dto.Name, floatData)
	alertExist, err := repository.Has(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
	}
	if alertExist {
		if err = repository.Update(ctx, dto.Name, alert); err != nil {
			return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	} else {
		if err = repository.Save(ctx, dto.Name, alert); err != nil {
			return nil, fmt.Errorf("failed save alert: %w", err)
		}
	}

	newAlert, errAlertNotFound := repository.Get(ctx, dto.Name)
	if errAlertNotFound != nil {
		return nil, fmt.Errorf(failedUpdateGaugeAlertPattern, errAlertNotFound)
	}

	return &newAlert, nil
}

func updateCounterAlert(ctx context.Context, dto dto.UpdateAlertDTO, repo UpdateStorage) (*entity.Alert, error) {
	intData, err := strconv.ParseInt(dto.Data, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(failedParseCounterValuePattern, err)
	}
	alert := entity.MakeCounterAlert(dto.Name, intData)
	hasAlert, err := repo.Has(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf("failed check alert exist: %w", err)
	}
	if !hasAlert {
		if err = repo.Save(ctx, dto.Name, alert); err != nil {
			return nil, fmt.Errorf("failed save gauge alert: %w", err)
		}

		resultAlert, err := repo.Get(ctx, dto.Name)
		if err != nil {
			return nil, fmt.Errorf("failed update counter alert: %w", err)
		}
		return &resultAlert, nil
	}
	oldAlert, err := repo.Get(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		if err = repo.Save(ctx, dto.Name, alert); err != nil {
			return nil, fmt.Errorf("failed save gauge alert: %w", err)
		}
		newAlert, err := repo.Get(ctx, dto.Name)
		if err != nil {
			return nil, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return &newAlert, nil
	}

	newValue := *oldAlert.IntValue + *alert.IntValue
	alert.IntValue = &newValue

	if err := repo.Update(ctx, dto.Name, alert); err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := repo.Get(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return &newAlert, nil
}
