package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"net/http"
)

func ShowHandler(strg storage.AlertStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		alert, err := strg.GetAlert(name)
		if err != nil {
			http.Error(w, "alert not found", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "name: %s, type: %s, value: %v", alert.Name, alert.Type, alert.Value)
	}
}
