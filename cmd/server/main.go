package main

import (
	"github.com/go-chi/chi/v5"
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
	router := chi.NewRouter()
	router.Get("/", handlers.DefaultHandler())
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Use(
			middleware.TypeValidator(),
			middleware.NameValidator(),
			middleware.ValueValidator(),
		)
		r.Post("/", handlers.UpdateHandler(repository))
	})
	return http.ListenAndServe(":8080", router)
}
