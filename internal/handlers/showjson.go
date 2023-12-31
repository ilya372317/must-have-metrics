package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

const (
	contentTypeHeader      = "content-type"
	jsonContentHeaderValue = "application/json"
)

type ShowJSONStorage interface {
	Get(ctx context.Context, name string) (entity.Alert, error)
}

func ShowJSONHandler(storage ShowJSONStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(contentTypeHeader, jsonContentHeaderValue)
		metrics, err := dto.NewMetricsDTOFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		showDTO := dto.CreateShowAlertDTOFromMetrics(metrics)
		if isValid, validErr := showDTO.Validate(); !isValid {
			http.Error(writer, validErr.Error(), http.StatusBadRequest)
			return
		}

		alert, err := storage.Get(request.Context(), showDTO.Name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		metrics = dto.NewMetricsDTOFromAlert(alert)
		response, err := json.Marshal(&metrics)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(response)
		if err != nil {
			logger.Log.Warn(err)
		}
	}
}
