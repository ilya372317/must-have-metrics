package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"net/http"
)

func AlertRouter(repository storage.AlertStorage, pathToFile string) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", handlers.IndexHandler(repository, pathToFile))
	router.Handle("/public/*", http.StripPrefix("/public", handlers.StaticHandler()))
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Use(
			middleware.TypeValidator(),
			middleware.NameValidator(),
			middleware.ValueValidator(),
		)
		r.Post("/", handlers.UpdateHandler(repository))
	})
	router.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Use(middleware.TypeValidator(), middleware.NameValidator())
		r.Get("/", handlers.ShowHandler(repository))
	})
	return router
}
