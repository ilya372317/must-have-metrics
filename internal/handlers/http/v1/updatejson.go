package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type updateJSONService interface {
	AddAlert(context.Context, dto.Metrics) (entity.Alert, error)
}

// UpdateJSONHandler allow to update specific metric by request in json format.
func UpdateJSONHandler(service updateJSONService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		metrics, err := dto.NewMetricsDTOFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if isValid, validErr := metrics.Validate(); !isValid {
			http.Error(writer, validErr.Error(), http.StatusBadRequest)
			return
		}
		newAlert, err := service.AddAlert(request.Context(), metrics)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			logger.Log.Warn(err)
			return
		}
		responseMetric := dto.NewMetricsDTOFromAlert(newAlert)
		response, err := json.Marshal(&responseMetric)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			logger.Log.Warn(err)
			return
		}
		if _, err = writer.Write(response); err != nil {
			logger.Log.Warn(err)
		}
	}
}
