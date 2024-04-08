package grpc

import (
	"context"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Show(ctx context.Context, in *pb.ShowMetricsRequest) (*pb.ShowMetricsResponse, error) {
	var response pb.ShowMetricsResponse

	showDTO := dto.ShowAlertDTO{
		Type: in.Type,
		Name: in.Id,
	}

	if ok, err := showDTO.Validate(); !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid grpc request: %v", err)
	}

	metrics, err := s.service.Get(ctx, showDTO.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "metrics with id [%s] not found: %v", showDTO.Name, err)
	}

	response.Metrics = &pb.Metrics{
		Value: metrics.FloatValue,
		Delta: metrics.IntValue,
		Id:    metrics.Name,
		Type:  metrics.Type,
	}

	return &response, nil
}
