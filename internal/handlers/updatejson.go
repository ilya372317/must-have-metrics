package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var zapLogger = logger.Get()

type UpdateJsonStorage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
}

func UpdateJsonHandler(storage UpdateJsonStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		metrics, err := dto.CreateMetricsDTOFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		updateDTO := dto.CreateUpdateAlertDTOFromMetrics(metrics)
		if isValid, validErr := updateDTO.Validate(); !isValid {
			http.Error(writer, validErr.Error(), http.StatusBadRequest)
			return
		}
		newAlert, err := service.AddAlert(storage, updateDTO)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			zapLogger.Error(err)
			return
		}
		responseMetric := dto.Metrics{
			ID:    newAlert.Name,
			MType: newAlert.Type,
		}
		if newAlert.Type == entity.TypeCounter {
			alertValue := newAlert.Value.(int64)
			responseMetric.Delta = &alertValue
		} else {
			alertValue := newAlert.Value.(float64)
			responseMetric.Value = &alertValue
		}
		response, err := json.Marshal(&responseMetric)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			zapLogger.Error(err)
			return
		}
		if _, err = writer.Write(response); err != nil {
			zapLogger.Error(err)
		}
	}
}
