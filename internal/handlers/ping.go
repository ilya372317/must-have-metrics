package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var pingLogger = logger.Get()

// PingHandler TODO: In 11 increment add ping function in new storage. And pass it in argument
// And delete getting DB from handler.
func PingHandler(serverConfig *config.ServerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		db, err := sql.Open("pgx", serverConfig.DatabaseDSN)
		if err != nil {
			http.Error(
				writer,
				fmt.Sprintf("Failed open connection to database: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		if pingErr := db.Ping(); pingErr != nil {
			http.Error(
				writer,
				fmt.Sprintf("Failed ping connection to database: %s", pingErr.Error()),
				http.StatusInternalServerError,
			)
		}
		if err = db.Close(); err != nil {
			pingLogger.Warnf("failed close connection with database: %v", err)
		}
		if _, err = fmt.Fprint(writer, "pong"); err != nil {
			pingLogger.Warnf("failed write data in response: %v", err)
		}
	}
}
