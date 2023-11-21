package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
)

type AlertStorage interface {
	Save(name string, alert entity.Alert)
	Update(name string, alert entity.Alert) error
	Get(name string) (entity.Alert, error)
	Has(name string) bool
	All() []entity.Alert
}

func AlertRouter(repository AlertStorage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.WithLogging())
	router.Get("/", handlers.IndexHandler(repository))
	router.Handle("/public/*", http.StripPrefix("/public", handlers.StaticHandler()))
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateJSONHandler(repository))
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ShowJSONHandler(repository))
	})
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Post("/", handlers.UpdateHandler(repository))
	})
	router.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Get("/", handlers.ShowHandler(repository))
	})
	return router
}
