package sender

import (
	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/client/cmiddleware"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
)

const failedSaveDataErrPattern = "failed to save data on server: %v\n"

type ReportSender func(agentConfig *config.AgentConfig, requestURL, body string)

func SendReport(agentConfig *config.AgentConfig, requestURL, body string) {
	c := resty.New()

	if agentConfig.ShouldSignData() {
		c.OnBeforeRequest(cmiddleware.WithSignature(agentConfig.SecretKey))
	}
	c.OnBeforeRequest(cmiddleware.WithCompress())

	_, err := c.R().SetBody(body).
		Post(requestURL)
	if err != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, err)
		return
	}
}
