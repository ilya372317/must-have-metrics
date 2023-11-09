package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/storage"
)

func ShowHandler(strg storage.Storage) http.HandlerFunc {
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
