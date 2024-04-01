package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	http2 "github.com/ilya372317/must-have-metrics/internal/handlers/http"
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

type MetricsService interface {
	AddAlert(context.Context, dto.Metrics) (entity.Alert, error)
	BulkAddAlerts(context.Context, []dto.Metrics) ([]entity.Alert, error)
	Ping() error
	GetAll(context.Context) ([]entity.Alert, error)
	Get(ctx context.Context, name string) (entity.Alert, error)
}

// AlertRouter return configured router.
func AlertRouter(service MetricsService, serverConfig *config.ServerConfig) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.WithLogging())
	if serverConfig.ShouldDecryptData() {
		router.Use(middleware.WithRSADecrypt(serverConfig.CryptoKey))
	}
	if serverConfig.ShouldCheckIP() {
		router.Use(middleware.WithTrustedSubnet(serverConfig))
	}
	router.Use(middleware.Compressed())
	router.Get("/", http2.IndexHandler(service))
	router.Get("/ping", http2.PingHandler(service))
	router.Handle("/public/*", http.StripPrefix("/public", http2.StaticHandler()))
	router.Route("/update", func(r chi.Router) {
		r.Post("/", http2.UpdateJSONHandler(service))
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(middleware.WithSign(serverConfig))
		r.Post("/", http2.BulkUpdate(service))
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", http2.ShowJSONHandler(service))
	})
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Post("/", http2.UpdateHandler(service))
	})
	router.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Get("/", http2.ShowHandler(service))
	})
	router.HandleFunc("/debug/pprof/*", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return router
}
