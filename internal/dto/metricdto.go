package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/dto/validator"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type MetricsList []Metrics

func NewMetricsListDTOFromRequest(r *http.Request) (MetricsList, error) {
	metricsList := make([]Metrics, 0)
	if err := json.NewDecoder(r.Body).Decode(&metricsList); err != nil {
		return nil, fmt.Errorf("failed create metrics list: %w", err)
	}

	return metricsList, nil
}

type Metrics struct {
	ID    string   `json:"id" valid:"type(string)"`          // Имя метрики
	MType string   `json:"type" valid:"in(gauge|counter)"`   // параметр, принимающий значение gauge или counter
	Value *float64 `json:"value,omitempty" valid:"optional"` // Значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty" valid:"optional"` // Значение метрики в случае передачи counter
}

func NewMetricsDTOFromRequest(r *http.Request) (Metrics, error) {
	metrics := Metrics{}
	defer func() {
		_ = r.Body.Close()
	}()
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		return metrics, fmt.Errorf("failed desirialize json body: %w", err)
	}

	return metrics, nil
}

func NewMetricsDTOFromRequestParams(r *http.Request) (*Metrics, error) {
	metrics := &Metrics{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "type"),
		Value: nil,
		Delta: nil,
	}
	value := chi.URLParam(r, "value")

	if metrics.MType == entity.TypeGauge {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed create metrics from request params: %w", err)
		}
		metrics.Value = &floatValue
	}

	if metrics.MType == entity.TypeCounter {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed create metrics from request params: %w", err)
		}
		metrics.Delta = &intValue
	}

	return metrics, nil
}

func NewMetricsDTOFromAlert(alert entity.Alert) Metrics {
	return Metrics{
		ID:    alert.Name,
		MType: alert.Type,
		Delta: alert.IntValue,
		Value: alert.FloatValue,
	}
}

func (dto *Metrics) Validate() (bool, error) {
	switch dto.MType {
	case entity.TypeGauge:
		if dto.Value == nil || dto.Delta != nil {
			return false, errors.New("gauge metric must have value field and must not have delta field")
		}
	case entity.TypeCounter:
		if dto.Delta == nil || dto.Value != nil {
			return false, errors.New("counter metric must have delta field and must not have value filed")
		}
	}

	isValid, err := validator.ValidateRequired(*dto)
	if err != nil {
		err = fmt.Errorf("metrics dto is invalid: %w", err)
	}

	return isValid, err
}
