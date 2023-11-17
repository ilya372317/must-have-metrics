package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/server/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type ShowStorage interface {
	Get(name string) (entity.Alert, error)
}

func ShowHandler(strg ShowStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showDTO := dto.ShowAlertDTOFromRequest(r)
		if _, err := showDTO.Validate(); err != nil {
			http.Error(w, fmt.Errorf("show parameters is invalid: %w", err).Error(), http.StatusBadRequest)
		}
		alert, err := strg.Get(showDTO.Name)
		if err != nil {
			http.Error(w, "alert not found", http.StatusNotFound)
			return
		}

		if _, err := fmt.Fprintf(w, "%v", alert.Value); err != nil {
			log.Println(err)
		}
	}
}
