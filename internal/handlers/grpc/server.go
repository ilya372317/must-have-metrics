package grpc

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	myinterceptor "github.com/ilya372317/must-have-metrics/internal/interceptor/server"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc"
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

func NewServer(cnfg *config.ServerConfig) *grpc.Server {
	return grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(logger.InterceptorLogger()),
		myinterceptor.WithTrustedSubnet(cnfg),
	))
}
