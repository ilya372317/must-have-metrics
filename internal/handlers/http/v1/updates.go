package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type bulkUpdateService interface {
	BulkAddAlerts(context.Context, []dto.Metrics) ([]entity.Alert, error)
}

// BulkUpdate allow to update multiply metrics by request in json format.
func BulkUpdate(service bulkUpdateService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		metricsList, err := dto.NewMetricsListDTOFromRequest(request)
		if err != nil {
			http.Error(writer, fmt.Sprintf("failed create metricsList dto: %v", err), http.StatusBadRequest)
			return
		}

		for _, metric := range metricsList {
			ok, err := metric.Validate()
			if !ok {
				http.Error(writer, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
			}
		}

		alerts, err := service.BulkAddAlerts(request.Context(), metricsList)

		if err != nil {
			http.Error(writer, fmt.Sprintf("failed insert metrics: %v", err), http.StatusInternalServerError)
			return
		}

		responseMetricsList := make([]dto.Metrics, 0, len(alerts))
		for _, alert := range alerts {
			responseMetricsList = append(responseMetricsList, dto.NewMetricsDTOFromAlert(alert))
		}
		response, err := json.Marshal(&responseMetricsList)
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
