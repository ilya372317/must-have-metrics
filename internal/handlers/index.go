package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/utils"
)

type IndexStorage interface {
	All() []entity.Alert
}

func IndexHandler(strg IndexStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		allAlerts := strg.All()
		sort.SliceStable(allAlerts, func(i, j int) bool {
			return allAlerts[i].Name < allAlerts[j].Name
		})
		tmpl, err := template.ParseFiles(utils.BasePath() + "/static" + "/index.html")
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
