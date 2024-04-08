package sender

import (
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/cmiddleware"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	pb "github.com/ilya372317/must-have-metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const failedSaveDataErrPattern = "failed to save data on server: %v\n"
const updateEndpoint = "updates"

// ReportData representation of collected metric for report.
type ReportData struct {
	Name  string
	Type  string
	Value uint64
	Delta int
}

// ReportSender interface for somehow sending report on server.
type ReportSender func(agentConfig *config.AgentConfig, data []ReportData)

// HTTPSendReport implementation of ReportSender interface wich send report on server by http request.
func HTTPSendReport(agentConfig *config.AgentConfig, data []ReportData) {
	c := resty.New()

	if agentConfig.ShouldSignData() {
		c.OnBeforeRequest(cmiddleware.WithSignature(agentConfig.SecretKey))
	}
	c.OnBeforeRequest(cmiddleware.WithRealIP())
	c.OnBeforeRequest(cmiddleware.WithCompress())
	if agentConfig.ShouldCipherData() {
		c.OnBeforeRequest(cmiddleware.WithRSACrypt(agentConfig.CryptoKey))
	}

	requestURL := "http://" + agentConfig.Host + "/" + updateEndpoint
	body := createJSONBody(data)
	_, err := c.R().SetBody(body).
		Post(requestURL)
	if err != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, err)
		return
	}
}

// GRPCSendReport implementation of ReportSender which send data on server by GRPC.
func GRPCSendReport(agentConfig *config.AgentConfig, data []ReportData) {
	ctx := context.Background()
	connect, err := grpc.DialContext(ctx,
		agentConfig.GRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Errorf("failed save data on server by grpc: %v", err)
		return
	}
	client := pb.NewMetricsServiceClient(connect)
	in := &pb.BulkUpdateMetricsRequest{}
	inMetrics := make([]*pb.Metrics, 0, len(data))
	for _, d := range data {
		var floatValuePtr *float64
		var intValuePtr *int64

		if d.Type == entity.TypeCounter {
			intValue := int64(d.Delta)
			intValuePtr = &intValue
		}

		if d.Type == entity.TypeGauge {
			floatValue := float64(d.Value)
			floatValuePtr = &floatValue
		}

		inMetrics = append(inMetrics, &pb.Metrics{
			Value: floatValuePtr,
			Delta: intValuePtr,
			Id:    d.Name,
			Type:  d.Type,
		})
	}
	in.Metrics = inMetrics

	if _, err = client.BulkUpdate(ctx, in); err != nil {
		logger.Log.Errorf("failed save data on server by GRPC: %v", err)
	}
}

func createJSONBody(data []ReportData) string {
	metricsList := make([]dto.Metrics, 0, len(data))
	for _, monitorValue := range data {
		m := dto.Metrics{
			ID:    monitorValue.Name,
			MType: monitorValue.Type,
		}
		if monitorValue.Type == entity.TypeCounter {
			int64Value := int64(monitorValue.Delta)
			m.Delta = &int64Value
		}
		if monitorValue.Type == entity.TypeGauge {
			float64Value := float64(monitorValue.Value)
			m.Value = &float64Value
		}
		metricsList = append(metricsList, m)
	}

	body, _ := json.Marshal(&metricsList)
	return string(body)
}
