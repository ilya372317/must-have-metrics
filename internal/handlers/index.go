package handlers

import (
	"fmt"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"html/template"
	"net/http"
	"sort"
)

func IndexHandler(strg storage.Storage, staticFolderPath string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		allAlerts := strg.All()
		sort.SliceStable(allAlerts, func(i, j int) bool {
			return allAlerts[i].Name < allAlerts[j].Name
		})
		tmpl, err := template.ParseFiles(staticFolderPath + "/index.html")
		if err != nil {
			http.Error(writer, fmt.Sprintf("internal server error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		if err = tmpl.Execute(writer, allAlerts); err != nil {
			http.Error(writer, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}
