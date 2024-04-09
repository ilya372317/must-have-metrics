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

type MetricsStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	All(ctx context.Context) ([]entity.Alert, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
	Ping() error
}

type MetricsService struct {
	storage MetricsStorage
	cnfg    *config.ServerConfig
}

func NewMetricsService(storage MetricsStorage, cnfg *config.ServerConfig) *MetricsService {
	return &MetricsService{storage: storage, cnfg: cnfg}
}

func (s *MetricsService) Get(ctx context.Context, id string) (entity.Alert, error) {
	alert, err := s.storage.Get(ctx, id)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("failed get metrics with id: %s, %w", id, err)
	}

	return alert, nil
}

func (s *MetricsService) GetAll(ctx context.Context) ([]entity.Alert, error) {
	result, err := s.storage.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get all metrics: %w", err)
	}
	return result, nil
}

func (s *MetricsService) Ping() error {
	err := s.storage.Ping()
	if err != nil {
		return fmt.Errorf("failed ping storage: %w", err)
	}
	return nil
}

// AddAlert save or update alert into storage.
func (s *MetricsService) AddAlert(
	ctx context.Context,
	dto dto.Metrics,
) (entity.Alert, error) {
	var alert entity.Alert
	var err error
	switch dto.MType {
	case entity.TypeGauge:
		alert, err = s.updateGaugeAlert(ctx, dto)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	case entity.TypeCounter:
		alert, err = s.updateCounterAlert(ctx, dto)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
		}
	default:
		return entity.Alert{}, errors.New("invalid type of metric")
	}

	if s.cnfg.StoreInterval == 0 {
		if err = StoreToFilesystem(ctx, s.storage, s.cnfg.FilePath); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save data to filesystem: %w", err)
		}
	}

	return alert, nil
}

func (s *MetricsService) updateGaugeAlert(ctx context.Context, dto dto.Metrics) (entity.Alert, error) {
	if dto.Value == nil {
		return entity.Alert{}, errors.New("failed update gauge alert, missing value field")
	}

	alert := entity.MakeGaugeAlert(dto.ID, *dto.Value)
	alertExist, err := s.storage.Has(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
	}
	if alertExist {
		if err = s.storage.Update(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, err)
		}
	} else {
		if err = s.storage.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save alert: %w", err)
		}
	}

	newAlert, errAlertNotFound := s.storage.Get(ctx, dto.ID)
	if errAlertNotFound != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateGaugeAlertPattern, errAlertNotFound)
	}

	return newAlert, nil
}

func (s *MetricsService) updateCounterAlert(ctx context.Context, dto dto.Metrics) (entity.Alert, error) {
	if dto.Delta == nil {
		return entity.Alert{}, errors.New("failed update counter alert, missing delta filed")
	}

	alert := entity.MakeCounterAlert(dto.ID, *dto.Delta)
	hasAlert, err := s.storage.Has(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("failed check alert exist: %w", err)
	}
	if !hasAlert {
		if err = s.storage.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save gauge alert: %w", err)
		}

		resultAlert, err := s.storage.Get(ctx, dto.ID)
		if err != nil {
			return entity.Alert{}, fmt.Errorf("failed update counter alert: %w", err)
		}
		return resultAlert, nil
	}
	oldAlert, err := s.storage.Get(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf("counter alert for update not found: %w", err)
	}

	if oldAlert.Type == entity.TypeGauge {
		if err = s.storage.Save(ctx, dto.ID, alert); err != nil {
			return entity.Alert{}, fmt.Errorf("failed save gauge alert: %w", err)
		}
		newAlert, err := s.storage.Get(ctx, dto.ID)
		if err != nil {
			return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
		}
		return newAlert, nil
	}

	newValue := *oldAlert.IntValue + *alert.IntValue
	alert.IntValue = &newValue

	if err := s.storage.Update(ctx, dto.ID, alert); err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	newAlert, err := s.storage.Get(ctx, dto.ID)
	if err != nil {
		return entity.Alert{}, fmt.Errorf(failedUpdateCounterPattern, err)
	}
	return newAlert, nil
}

// BulkAddAlerts saving multiply given alerts into storage.
func (s *MetricsService) BulkAddAlerts(ctx context.Context, metricsList []dto.Metrics) ([]entity.Alert, error) {
	ids := make([]string, 0, len(metricsList))
	for _, metrics := range metricsList {
		ids = append(ids, metrics.ID)
	}

	if err := s.bulkAddGaugeAlerts(ctx, metricsList); err != nil {
		return nil, fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	if err := s.bulkAddCounterAlerts(ctx, metricsList); err != nil {
		return nil, fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	resultAlerts, err := s.storage.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed get alerts by ids: %w", err)
	}

	return resultAlerts, nil
}

func (s *MetricsService) bulkAddGaugeAlerts(ctx context.Context, metricsList dto.MetricsList) error {
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

	if err := s.storage.BulkInsertOrUpdate(ctx, alerts); err != nil {
		return fmt.Errorf(failedBulkAddAlertsErrPattern, err)
	}

	return nil
}

func (s *MetricsService) bulkAddCounterAlerts(ctx context.Context, metricsList dto.MetricsList) error {
	counterMetrics := make([]dto.Metrics, 0, len(metricsList))
	for _, metrics := range metricsList {
		if metrics.MType == entity.TypeCounter {
			counterMetrics = append(counterMetrics, metrics)
		}
	}
	for _, metrics := range counterMetrics {
		if _, err := s.updateCounterAlert(ctx, metrics); err != nil {
			return fmt.Errorf(failedBulkAddAlertsErrPattern, err)
		}
	}

	return nil
}
