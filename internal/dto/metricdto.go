package dto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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
		value, ok := alert.Value.(int64)
		if !ok {
			value = int64(alert.Value.(float64))
		}
		result.Delta = &value
	} else {
		value, ok := alert.Value.(float64)
		if !ok {
			value = float64(alert.Value.(int64))
		}
		result.Value = &value
	}

	return result
}
