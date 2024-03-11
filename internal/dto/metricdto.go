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

// NewMetricsListDTOFromRequest create MetricsList from given request
func NewMetricsListDTOFromRequest(r *http.Request) (MetricsList, error) {
	metricsList := make([]Metrics, 0)
	if err := json.NewDecoder(r.Body).Decode(&metricsList); err != nil {
		return nil, fmt.Errorf("failed create metrics list: %w", err)
	}

	return metricsList, nil
}

// Metrics DTO for representing received metrics from agent.
type Metrics struct {
	Delta *int64   `json:"delta,omitempty" valid:"optional"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty" valid:"optional"` // Значение метрики в случае передачи gauge
	ID    string   `json:"id" valid:"type(string)"`          // Имя метрики
	MType string   `json:"type" valid:"in(gauge|counter)"`   // параметр, принимающий значение gauge или counter
}

// NewMetricsDTOFromRequest create Metrics DTO from given request.
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

// NewMetricsDTOFromRequestParams create Metrics from given request.
// For build using query parameters instead of request body.
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

// NewMetricsDTOFromAlert create Metrics from given entity.Alert.
func NewMetricsDTOFromAlert(alert entity.Alert) Metrics {
	return Metrics{
		ID:    alert.Name,
		MType: alert.Type,
		Delta: alert.IntValue,
		Value: alert.FloatValue,
	}
}

// ConvertToAlert converting Metrics to entity.Alert.
func (dto *Metrics) ConvertToAlert() entity.Alert {
	alert := entity.Alert{
		Type: dto.MType,
		Name: dto.ID,
	}

	if dto.MType == entity.TypeGauge {
		alert.FloatValue = dto.Value
	}
	if dto.MType == entity.TypeCounter {
		alert.IntValue = dto.Delta
	}

	return alert
}

// Validate perform validation on Metrics
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
