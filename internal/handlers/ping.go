package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var pingLogger = logger.Get()

// PingHandler TODO: In 11 increment add ping function in new storage. And pass it in argument
// And delete getting DB from handler.
func PingHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			`localhost`, `ilya`, `Ilya372317`, `metrics`)
		db, err := sql.Open("pgx", ps)
		if err != nil {
			http.Error(
				writer,
				fmt.Sprintf("Failed get connection to database: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
		if err = db.Close(); err != nil {
			pingLogger.Warnf("failed close connection with database: %v", err)
		}
		if _, err = fmt.Fprint(writer, "pong"); err != nil {
			pingLogger.Warnf("failed write data in response: %v", err)
		}
	}
}
