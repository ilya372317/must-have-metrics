package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
)

func CreateSign(body []byte, secretKey string) []byte {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(body)
	return h.Sum(nil)
}

func IsCorrectSigned(serverConfig *config.ServerConfig, request *http.Request) (bool, error) {
	if !serverConfig.ShouldSignData() {
		return true, nil
	}

	agentSign := request.Header.Get("HashSHA256")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return false, fmt.Errorf("failed read request body for check sign: %w", err)
	}
	sign := CreateSign(body, serverConfig.SecretKey)
	encodeSign := base64.StdEncoding.EncodeToString(sign)
	if agentSign != encodeSign {
		return false, nil
	}

	if err := request.Body.Close(); err != nil {
		logger.Log.Warnf("failed close body in signature middleware: %v", err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	return true, nil
}

func SetSign(writer http.ResponseWriter, serverConfig *config.ServerConfig, data []byte) {
	if !serverConfig.ShouldSignData() {
		return
	}
	sign := CreateSign(data, serverConfig.SecretKey)
	encodeSign := base64.StdEncoding.EncodeToString(sign)
	writer.Header().Set("HashSHA256", encodeSign)
}
