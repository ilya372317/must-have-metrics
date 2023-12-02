package handlers

import (
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
)

type UpdateStorage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
	AllWithKeys() map[string]entity.Alert
	Fill(map[string]entity.Alert)
}

func UpdateHandler(storage UpdateStorage, serverConfig *config.ServerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		updateAlertDTO := dto.CreateUpdateAlertDTOFromRequest(request)
		_, err := updateAlertDTO.Validate()
		if err != nil {
			http.Error(writer, fmt.Errorf("invalid parameters: %w", err).Error(), http.StatusBadRequest)
		}
		if _, err := service.AddAlert(storage, updateAlertDTO, serverConfig); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}
}
