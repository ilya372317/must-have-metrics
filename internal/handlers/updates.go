package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
)

type BulkUpdateStorage interface {
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
}

func BulkUpdate(storage BulkUpdateStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		isSignCorrect, err := isCorrectSigned(serverConfig, request)
		if err != nil {
			http.Error(writer, fmt.Sprintf("failed check sign: %v", err), http.StatusInternalServerError)
			return
		}
		if !isSignCorrect {
			http.Error(writer, "invalid sign", http.StatusBadRequest)
			return
		}
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
		setSign(writer, serverConfig, response)
		if _, err = writer.Write(response); err != nil {
			logger.Log.Warn(err)
		}
	}
}
