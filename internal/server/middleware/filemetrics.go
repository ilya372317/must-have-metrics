package middleware

import (
	"net/http"
	"strconv"

	"github.com/ilya372317/must-have-metrics/internal/config"
)

type FilesystemSupportStorage interface {
	StoreToFilesystem(filepath string) error
}

func SaveMetricsInFile(repo FilesystemSupportStorage) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			serverConfig, err := config.GetServerConfig()
			if err != nil {
				http.Error(writer, "failed init server configuration", http.StatusInternalServerError)
				return
			}

			storeInterval, err := strconv.Atoi(serverConfig.GetValue(config.StoreInterval))
			if err != nil {
				http.Error(writer, "invalid store interval config value", http.StatusInternalServerError)
				return
			}
			isSync := storeInterval > 0

			handler.ServeHTTP(writer, request)

			if !isSync {
				return
			}

			if err = repo.StoreToFilesystem(serverConfig.GetValue(config.StorePath)); err != nil {
				http.Error(writer, "Failed store metrics in filesystem", http.StatusInternalServerError)
			}
		})
	}
}
