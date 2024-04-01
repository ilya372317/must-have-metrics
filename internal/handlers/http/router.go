package http

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/handlers/http/v1"
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

type MetricsRouter struct {
	cnfg    *config.ServerConfig
	service MetricsService
}

func NewMetricsRouter(cnfg *config.ServerConfig, service MetricsService) *MetricsRouter {
	return &MetricsRouter{
		cnfg:    cnfg,
		service: service,
	}
}

// BuildRouter return configured router.
func (rt *MetricsRouter) BuildRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.WithLogging())
	if rt.cnfg.ShouldDecryptData() {
		router.Use(middleware.WithRSADecrypt(rt.cnfg.CryptoKey))
	}
	if rt.cnfg.ShouldCheckIP() {
		router.Use(middleware.WithTrustedSubnet(rt.cnfg))
	}
	router.Use(middleware.Compressed())
	router.Get("/", v1.IndexHandler(rt.service))
	router.Get("/ping", v1.PingHandler(rt.service))
	router.Handle("/public/*", http.StripPrefix("/public", v1.StaticHandler()))
	router.Route("/update", func(r chi.Router) {
		r.Post("/", v1.UpdateJSONHandler(rt.service))
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(middleware.WithSign(rt.cnfg))
		r.Post("/", v1.BulkUpdate(rt.service))
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", v1.ShowJSONHandler(rt.service))
	})
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Post("/", v1.UpdateHandler(rt.service))
	})
	router.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Get("/", v1.ShowHandler(rt.service))
	})
	router.HandleFunc("/debug/pprof/*", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return router
}
