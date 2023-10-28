package main

import (
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
	storage2 "github.com/ilya372317/must-have-metrics/internal/storage"
	"log"
	"net/http"
)

var storage storage2.AlertStorage

func init() {
	storageValue := storage2.MakeAlertInMemoryStorage()
	storage = &storageValue
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
			handlers.UpdateHandler(storage),
			middleware.ValidUpdate(),
			middleware.Method(http.MethodPost),
		),
	)
	return http.ListenAndServe(":8080", mux)
}
