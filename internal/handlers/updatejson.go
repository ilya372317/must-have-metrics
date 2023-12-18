package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
)

type UpdateJSONStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
}

func UpdateJSONHandler(storage UpdateJSONStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
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
		newAlert, err := service.AddAlert(request.Context(), storage, metrics, serverConfig)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			logger.Log.Warn(err)
			return
		}
		responseMetric := dto.NewMetricsDTOFromAlert(*newAlert)
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
