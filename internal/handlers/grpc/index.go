package grpc

import (
	"context"

	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Index(ctx context.Context, _ *pb.IndexMetricsRequest) (*pb.IndexMetricsResponse, error) {
	metrics, err := s.service.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get all metrics by grpc: %s", err)
	}

	outMetrics := make([]*pb.Metrics, 0, len(metrics))

	for _, m := range metrics {
		outMetrics = append(outMetrics, &pb.Metrics{
			Value: m.FloatValue,
			Delta: m.IntValue,
			Id:    m.Name,
			Type:  m.Type,
		})
	}
	var response pb.IndexMetricsResponse
	response.Metrics = outMetrics

	return &response, nil
}
