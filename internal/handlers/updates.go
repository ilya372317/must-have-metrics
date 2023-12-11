package handlers

import (
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
)

func BulkUpdate(storage UpdateJSONStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
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
	}
}
