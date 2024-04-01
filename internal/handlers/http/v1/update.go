package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type updateService interface {
	AddAlert(context.Context, dto.Metrics) (entity.Alert, error)
}

// UpdateHandler allow update specific metric by request in plain text format.
func UpdateHandler(service updateService) http.HandlerFunc {
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
		if _, err := service.AddAlert(request.Context(), *metrics); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}
