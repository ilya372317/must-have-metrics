package middleware

import "net/http"

type FilesystemSupportStorage interface {
	StoreToFilesystem(filepath string) error
}

func SavingMetricsInFile(repo FilesystemSupportStorage, filepath string) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			handler.ServeHTTP(writer, request)

			if err := repo.StoreToFilesystem(filepath); err != nil {
				http.Error(writer, "Failed store metrics in filesystem", http.StatusInternalServerError)
			}
		})
	}
}
