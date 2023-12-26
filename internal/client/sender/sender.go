package sender

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/signature"
	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
)

const failedSaveDataErrPattern = "failed to save data on server: %v\n"

type Sender struct {
	Client      *SenderClient
	agentConfig *config.AgentConfig
}

type SenderClient struct {
	*http.Client
	agentConfig *config.AgentConfig
}

func NewSender(client *SenderClient, agentConfig *config.AgentConfig) *Sender {
	return &Sender{
		Client:      client,
		agentConfig: agentConfig,
	}
}

func (c *SenderClient) Do(req *http.Request) (*http.Response, error) {
	c.agentConfig.ShouldSignData()
	{
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed read body for sign: %w", err)
		}

		sign := signature.CreateSign(body, c.agentConfig.SecretKey)
		encodeSing := base64.StdEncoding.EncodeToString(sign)
		req.Header.Set("HashSHA256", encodeSing)

		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return c.Client.Do(req)
}

func (s *Sender) Send(body string) {
	compressedData, errCompress := compress.Do([]byte(body))
	if errCompress != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, errCompress)
	}
	requestURL := createURLForReportStat(s.agentConfig.Host)
	request, errRequest := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(compressedData))
	if errRequest != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, errRequest)
		return
	}
	request.Header.Set("Content-Encoding", "gzip")

	res, err := s.Client.Do(request)
	if err != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, err)
		return
	}
	_ = res.Body.Close()
}

func createURLForReportStat(host string) string {
	return fmt.Sprintf("http://" + host + "/updates")
}

func NewSenderClient(agentConfig *config.AgentConfig, client *http.Client) *SenderClient {
	return &SenderClient{
		Client:      client,
		agentConfig: agentConfig,
	}
}
