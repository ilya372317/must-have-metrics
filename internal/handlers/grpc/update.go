package grpc

import (
	"context"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Update(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse

	inMetrics := in.Metrics
	updateDTO := dto.Metrics{
		Delta: inMetrics.Delta,
		Value: inMetrics.Value,
		ID:    inMetrics.Id,
		MType: inMetrics.Type,
	}

	if ok, err := updateDTO.Validate(); !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument for single update from grpc: %v", err)
	}

	updatedMetrics, err := s.service.AddAlert(ctx, updateDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed add or update alert from grpc: %v", err)
	}

	response.Metrics = &pb.Metrics{
		Value: updatedMetrics.FloatValue,
		Delta: updatedMetrics.IntValue,
		Id:    updatedMetrics.Name,
		Type:  updatedMetrics.Type,
	}

	return &response, nil
}
