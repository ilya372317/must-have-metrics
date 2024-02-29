package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/handlers"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/middleware"
)

// AlertStorage interface with all storage methods. Different handler will use different methods from here.
type AlertStorage interface {
	Save(ctx context.Context, name string, alert entity.Alert) error
	Update(ctx context.Context, name string, alert entity.Alert) error
	Get(ctx context.Context, name string) (entity.Alert, error)
	Has(ctx context.Context, name string) (bool, error)
	All(ctx context.Context) ([]entity.Alert, error)
	AllWithKeys(ctx context.Context) (map[string]entity.Alert, error)
	Fill(context.Context, map[string]entity.Alert) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Alert, error)
	BulkInsertOrUpdate(ctx context.Context, alerts []entity.Alert) error
	Ping() error
}

// AlertRouter return configured router.
func AlertRouter(repository AlertStorage, serverConfig *config.ServerConfig) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.WithLogging(), middleware.Compressed())
	router.Get("/", handlers.IndexHandler(repository))
	router.Get("/ping", handlers.PingHandler(repository))
	router.Handle("/public/*", http.StripPrefix("/public", handlers.StaticHandler()))
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateJSONHandler(repository, serverConfig))
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(middleware.WithSign(serverConfig))
		r.Post("/", handlers.BulkUpdate(repository))
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
	router.HandleFunc("/debug/pprof/*", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return router
}
