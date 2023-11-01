package handlers

import (
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"html/template"
	"net/http"
	"sort"
)

func IndexHandler(strg storage.AlertStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		allAlerts := strg.AllAlert()
		sort.SliceStable(allAlerts, func(i, j int) bool {
			return allAlerts[i].Name < allAlerts[j].Name
		})
		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			http.Error(writer, "internal server error", http.StatusInternalServerError)
			return
		}
		if err = tmpl.Execute(writer, allAlerts); err != nil {
			http.Error(writer, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}
