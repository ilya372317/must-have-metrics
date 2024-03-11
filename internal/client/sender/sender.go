package sender

import (
	"github.com/go-resty/resty/v2"
	cmiddleware2 "github.com/ilya372317/must-have-metrics/internal/cmiddleware"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
)

const failedSaveDataErrPattern = "failed to save data on server: %v\n"

// ReportSender interface for somehow sending report on server.
type ReportSender func(agentConfig *config.AgentConfig, requestURL, body string)

// SendReport implementation of ReportSender interface wich send report on server by http request.
func SendReport(agentConfig *config.AgentConfig, requestURL, body string) {
	c := resty.New()

	if agentConfig.ShouldSignData() {
		c.OnBeforeRequest(cmiddleware2.WithSignature(agentConfig.SecretKey))
	}
	c.OnBeforeRequest(cmiddleware2.WithCompress())

	_, err := c.R().SetBody(body).
		Post(requestURL)
	if err != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, err)
		return
	}
}
