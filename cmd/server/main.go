package main

import (
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"github.com/ilya372317/must-have-metrics/internal/router"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"log"
	"net/http"
)

var repository storage.AlertStorage

func init() {
	repository = storage.MakeInMemoryStorage()
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("vailed to start server on port 8080: %v", err)
	}
}

func run() error {
	return http.ListenAndServe(":8080", router.AlertRouter(repository, constant.StaticFilePath))
}
