package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
)

type ShowStorage interface {
	Get(name string) (entity.Alert, error)
}

func ShowHandler(strg ShowStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		alert, err := strg.Get(name)
		if err != nil {
			http.Error(w, "alert not found", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "%v", alert.Value)
	}
}
