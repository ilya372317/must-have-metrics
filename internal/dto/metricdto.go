package dto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type Metrics struct { //nolint:govet // i don`t now how to fix it
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
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
	if err = json.Unmarshal(body, &metrics); err != nil {
		return metrics, fmt.Errorf("failed desirialize json body: %w", err)
	}

	return metrics, nil
}

func CreateMetricsDTOFromAlert(alert entity.Alert) Metrics {
	result := Metrics{
		ID:    alert.Name,
		MType: alert.Type,
	}
	if alert.Type == entity.TypeCounter {
		value, _ := alert.Value.(int64)
		result.Delta = &value
	} else {
		value, _ := alert.Value.(float64)
		result.Value = &value
	}

	return result
}
