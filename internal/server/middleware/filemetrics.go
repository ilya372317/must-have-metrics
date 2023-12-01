package middleware

import (
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
)

type FilesystemSupportStorage interface {
	StoreToFilesystem(filepath string) error
}

func SaveMetricsInFile(repo FilesystemSupportStorage, cnfg *config.Config) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			isSync := cnfg.StoreInterval > 0

			handler.ServeHTTP(writer, request)

			if !isSync {
				return
			}

			if err := repo.StoreToFilesystem(cnfg.FilePath); err != nil {
				http.Error(writer, "Failed store metrics in filesystem", http.StatusInternalServerError)
			}
		})
	}
}
