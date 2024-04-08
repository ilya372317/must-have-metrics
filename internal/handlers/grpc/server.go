package grpc

import (
	"context"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	pb "github.com/ilya372317/must-have-metrics/proto"
)

type metricsService interface {
	AddAlert(context.Context, dto.Metrics) (entity.Alert, error)
	BulkAddAlerts(context.Context, []dto.Metrics) ([]entity.Alert, error)
	Ping() error
	GetAll(context.Context) ([]entity.Alert, error)
	Get(ctx context.Context, name string) (entity.Alert, error)
}

type Server struct {
	pb.UnimplementedMetricsServiceServer

	service metricsService
	cnfg    *config.ServerConfig
}

func New(service metricsService, cnfg *config.ServerConfig) *Server {
	return &Server{
		service: service,
		cnfg:    cnfg,
	}
}
