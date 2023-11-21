package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var showLogger = logger.Get()

type ShowJSONStorage interface {
	Get(name string) (entity.Alert, error)
}

func ShowJSONHandler(storage ShowJSONStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		metrics, err := dto.CreateMetricsDTOFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		showDTO := dto.CreateShowAlertDTOFromMetrics(metrics)
		if isValid, validErr := showDTO.Validate(); !isValid {
			http.Error(writer, validErr.Error(), http.StatusBadRequest)
			return
		}

		alert, err := storage.Get(showDTO.Name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		metrics = dto.CreateMetricsDTOFromAlert(alert)
		response, err := json.Marshal(&metrics)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(response)
		if err != nil {
			showLogger.Error(err)
		}
	}
}
