package http

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/utils"
)

type indexService interface {
	GetAll(ctx context.Context) ([]entity.Alert, error)
}

// IndexHandler give list of stored metrics in html format.
func IndexHandler(service indexService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		allAlerts, err := service.GetAll(request.Context())
		if err != nil {
			http.Error(writer,
				fmt.Sprintf("failed get data from storage: %v", err), http.StatusInternalServerError)
			return
		}
		sort.SliceStable(allAlerts, func(i, j int) bool {
			return allAlerts[i].Name < allAlerts[j].Name
		})
		tmpl, err := template.ParseFiles(utils.Root + "/static" + "/index.html")
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
