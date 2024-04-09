package grpc

import (
	"context"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) BulkUpdate(
	ctx context.Context,
	in *pb.BulkUpdateMetricsRequest,
) (*pb.BulkUpdateMetricsResponse, error) {
	var response pb.BulkUpdateMetricsResponse
	metrics := make([]dto.Metrics, 0, len(in.Metrics))

	for _, inMetric := range in.Metrics {
		newMetric := dto.Metrics{
			Delta: inMetric.Delta,
			Value: inMetric.Value,
			ID:    inMetric.Id,
			MType: inMetric.Type,
		}
		isValid, err := newMetric.Validate()
		if !isValid {
			return nil, status.Errorf(codes.InvalidArgument, "invalid request data given: %v", err)
		}

		metrics = append(metrics, newMetric)
	}

	alerts, err := s.service.BulkAddAlerts(ctx, metrics)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed bulk store alerts by grpc: %v", err)
	}

	responseMetrics := make([]*pb.Metrics, 0, len(alerts))

	for _, a := range alerts {
		responseMetrics = append(responseMetrics, &pb.Metrics{
			Id:    a.Name,
			Type:  a.Type,
			Delta: a.IntValue,
			Value: a.FloatValue,
		})
	}

	response.Metrics = responseMetrics

	return &response, nil
}
