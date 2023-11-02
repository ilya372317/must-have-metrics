package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
	"github.com/ilya372317/must-have-metrics/internal/storage"
)

func AlertRouter(repository storage.AlertStorage) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", handlers.IndexHandler(repository, constant.StaticFilePath))
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
