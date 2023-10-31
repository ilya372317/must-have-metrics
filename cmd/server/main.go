package main

import (
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.DefaultHandler())
	mux.HandleFunc(
		"/update/",
		middleware.Chain(
			handlers.UpdateHandler(repository),
			middleware.ValidUpdate(),
			middleware.Method(http.MethodPost),
		),
	)
	return http.ListenAndServe(":8080", mux)
}
