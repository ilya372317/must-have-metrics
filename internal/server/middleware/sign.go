package middleware

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/signature"
)

type signWriter struct {
	http.ResponseWriter
	serverConfig *config.ServerConfig
}

func (s signWriter) Write(data []byte) (int, error) {
	setSign(s, s.serverConfig, data)
	return s.ResponseWriter.Write(data)
}

func setSign(writer http.ResponseWriter, serverConfig *config.ServerConfig, data []byte) {
	sign := signature.CreateSign(data, serverConfig.SecretKey)
	encodeSign := base64.StdEncoding.EncodeToString(sign)
	writer.Header().Set("HashSHA256", encodeSign)
}

func isCorrectSigned(serverConfig *config.ServerConfig, request *http.Request) (bool, error) {
	agentSign := request.Header.Get("HashSHA256")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return false, fmt.Errorf("failed read request body for check sign: %w", err)
	}
	sign := signature.CreateSign(body, serverConfig.SecretKey)
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

// WithSign add sign to endpoint based on it response.
func WithSign(serverConfig *config.ServerConfig) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			sw := writer
			if serverConfig.ShouldSignData() {
				sw = signWriter{
					ResponseWriter: writer,
					serverConfig:   serverConfig,
				}

				correctSigned, err := isCorrectSigned(serverConfig, request)
				if err != nil {
					http.Error(writer, fmt.Sprintf("failed check sign: %v", err), http.StatusInternalServerError)
					return
				}
				if !correctSigned {
					http.Error(writer, "invalid sign", http.StatusBadRequest)
					return
				}
			}
			handler.ServeHTTP(sw, request)
		})
	}
}
