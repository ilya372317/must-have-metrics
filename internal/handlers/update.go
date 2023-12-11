package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
)

type UpdateStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
}

func UpdateHandler(storage UpdateStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metrics, err := dto.NewMetricsDTOFromRequestParams(request)
		if err != nil {
			http.Error(writer, fmt.Sprintf("invalid request parameters: %v", err), http.StatusBadRequest)
			return
		}

		ok, err := metrics.Validate()
		if !ok {
			http.Error(writer, fmt.Errorf("invalid parameters: %w", err).Error(), http.StatusBadRequest)
		}
		if _, err := service.AddAlert(request.Context(), storage, *metrics, serverConfig); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}
