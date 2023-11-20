package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

var errMetricValueIsInvalid = errors.New("metric value is invalid")

type Metrics struct {
	ID    string   `json:"id"`                             // имя метрики
	MType string   `json:"type" valid:"in(gauge|counter)"` // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"`                // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"`                // значение метрики в случае передачи gauge
}

func CreateMetricsDTOFromRequest(r *http.Request) (Metrics, error) {
	metrics := Metrics{}
	defer func() {
		_ = r.Body.Close()
	}()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return metrics, fmt.Errorf("failed read body: %w", err)
	}
	if err := json.Unmarshal(body, &metrics); err != nil {
		return metrics, fmt.Errorf("failed desirialize json body: %w", err)
	}

	return metrics, err
}

func CreateMetricsDTOFromAlert(alert entity.Alert) Metrics {
	result := Metrics{
		ID:    alert.Name,
		MType: alert.Type,
	}
	if alert.Type == entity.TypeCounter {
		alertValue := alert.Value.(int64)
		result.Delta = &alertValue
	} else {
		alertValue := alert.Value.(float64)
		result.Value = &alertValue
	}

	return result
}

func (m *Metrics) UnmarshalJSON(data []byte) error {
	type metricAlias Metrics
	metric := metricAlias{}
	if err := json.Unmarshal(data, &metric); err != nil {
		return err
	}
	*m = Metrics(metric)

	switch m.MType {
	case entity.TypeCounter:
		if m.Delta == nil || m.Value != nil {
			return errMetricValueIsInvalid
		}
		break
	case entity.TypeGauge:
		if m.Delta != nil || m.Value == nil {
			return errMetricValueIsInvalid
		}
		break
	default:
		return errors.New("metrics type not define or invalid")
	}
	return nil
}
