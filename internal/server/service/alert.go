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
const (
	failedUpdateGaugeAlertPattern = "failed update gauge alert: %w"
	failedBulkAddAlertsErrPattern = "failed bulk insert alerts: %w"
)

type updateStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
}

type bulkUpdateStorage interface {
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
}

type updateCounterStorage interface {
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
}

// AddAlert save or update alert into storage.
func AddAlert(
	ctx context.Context,
	repo updateStorage,
	dto dto.Metrics,
	serverConfig *config.ServerConfig,
) (entity.Alert, error) {
	var alert entity.Alert
	var err error
	switch dto.MType {
	case entity.TypeGauge:
		alert, err = updateGaugeAlert(ctx, dto, repo)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	case entity.TypeCounter:
		alert, err = updateCounterAlert(ctx, dto, repo)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
		}
	default:
		return entity.Alert{}, errors.New("invalid type of metric")
	}

	if serverConfig.StoreInterval == 0 {
		if err = StoreToFilesystem(ctx, repo, serverConfig.FilePath); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save data to filesystem: %w", err)
		}
	}

	return alert, nil
}

func updateGaugeAlert(ctx context.Context, dto dto.Metrics, repository updateStorage) (entity.Alert, error) {
	if dto.Value == nil {
		return entity.Alert{}, errors.New("failed update gauge alert, missing value field")
	}

	alert := entity.MakeGaugeAlert(dto.ID, *dto.Value)
	alertExist, err := repository.Has(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
	}
	if alertExist {
		if err = repository.Update(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	} else {
		if err = repository.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save alert: %w", err)
		}
	}

	newAlert, errAlertNotFound := repository.Get(ctx, dto.ID)
	if errAlertNotFound != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, errAlertNotFound)
	}

	return newAlert, nil
}

func updateCounterAlert(ctx context.Context, dto dto.Metrics, repo updateCounterStorage) (entity.Alert, error) {
	if dto.Delta == nil {
		return entity.Alert{}, errors.New("failed update counter alert, missing delta filed")
	}

	alert := entity.MakeCounterAlert(dto.ID, *dto.Delta)
	hasAlert, err := repo.Has(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("failed check alert exist: %w", err)
	}
	if !hasAlert {
		if err = repo.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save gauge alert: %w", err)
		}

		resultAlert, err := repo.Get(ctx, dto.ID)
		if err != nil {
			return entity.Alert{}, fmt.Errorf("failed update counter alert: %w", err)
		}
		return resultAlert, nil
	}
	oldAlert, err := repo.Get(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		if err = repo.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save gauge alert: %w", err)
		}
		newAlert, err := repo.Get(ctx, dto.ID)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return newAlert, nil
	}

	newValue := *oldAlert.IntValue + *alert.IntValue
	alert.IntValue = &newValue

	if err := repo.Update(ctx, dto.ID, alert); err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := repo.Get(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return newAlert, nil
}

// BulkAddAlerts saving multiply given alerts into storage.
func BulkAddAlerts(ctx context.Context, storage bulkUpdateStorage, metricsList []dto.Metrics) ([]entity.Alert, error) {
	ids := make([]string, 0, len(metricsList))
	for _, metrics := range metricsList {
		ids = append(ids, metrics.ID)
	}

	if err := bulkAddGaugeAlerts(ctx, storage, metricsList); err != nil {
		return nil, fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	if err := bulkAddCounterAlerts(ctx, storage, metricsList); err != nil {
		return nil, fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	resultAlerts, err := storage.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed get alerts by ids: %w", err)
	}

	return resultAlerts, nil
}

func bulkAddGaugeAlerts(ctx context.Context,
	storage bulkUpdateStorage, metricsList dto.MetricsList) error {
	alerts := make([]entity.Alert, 0, len(metricsList))
	gaugeMetrics := make([]dto.Metrics, 0, len(metricsList))
	for _, metrics := range metricsList {
		if metrics.MType == entity.TypeGauge {
			gaugeMetrics = append(gaugeMetrics, metrics)
		}
	}
	for _, metrics := range gaugeMetrics {
		alerts = append(alerts, metrics.ConvertToAlert())
	}

	if err := storage.BulkInsertOrUpdate(ctx, alerts); err != nil {
		return fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	return nil
}

func bulkAddCounterAlerts(ctx context.Context,
	storage bulkUpdateStorage, metricsList dto.MetricsList) error {
	counterMetrics := make([]dto.Metrics, 0, len(metricsList))
	for _, metrics := range metricsList {
		if metrics.MType == entity.TypeCounter {
			counterMetrics = append(counterMetrics, metrics)
		}
	}
	for _, metrics := range counterMetrics {
		if _, err := updateCounterAlert(ctx, metrics, storage); err != nil {
			return fmt.Errorf(failedBulkAddAlertsErrPattern, err)
		}
	}

	return nil
}
