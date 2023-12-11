package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
)

type AlertStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	All(ctx context.Context) ([]entity.Alert, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
}

func AlertRouter(repository AlertStorage, serverConfig *config.ServerConfig) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.WithLogging())
	router.Use(middleware.Compressed())
	router.Get("/", handlers.IndexHandler(repository))
	router.Get("/ping", handlers.PingHandler(serverConfig))
	router.Handle("/public/*", http.StripPrefix("/public", handlers.StaticHandler()))
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateJSONHandler(repository, serverConfig))
	})
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", handlers.BulkUpdate(repository, serverConfig))
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ShowJSONHandler(repository))
	})
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Post("/", handlers.UpdateHandler(repository, serverConfig))
	})
	router.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Get("/", handlers.ShowHandler(repository))
	})
	return router
}
