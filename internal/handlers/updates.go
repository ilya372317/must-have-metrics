package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
)

type BulkSupportStorage interface {
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
}

func BulkUpdate(storage BulkSupportStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
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

		alerts, err := service.BulkAddAlerts(request.Context(), storage, metricsList)

		responseMetricsList := make([]dto.Metrics, 0, len(alerts))
		for _, alert := range alerts {
			responseMetricsList = append(responseMetricsList, dto.NewMetricsDTOFromAlert(alert))
		}
		response, err := json.Marshal(&responseMetricsList)
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
