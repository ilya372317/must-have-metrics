package handlers

import (
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type pingStorage interface {
	Ping() error
}

// PingHandler allow to check connection with storage.
func PingHandler(repository pingStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if pingErr := repository.Ping(); pingErr != nil {
			http.Error(
				writer,
				fmt.Sprintf("Failed ping connection to database: %s", pingErr.Error()),
				http.StatusInternalServerError,
			)
		}
		if _, err := fmt.Fprint(writer, "pong"); err != nil {
			logger.Log.Warnf("failed write data in response: %v", err)
		}
	}
}
