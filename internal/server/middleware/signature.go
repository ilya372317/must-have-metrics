package middleware

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

func Signature(serverConfig *config.ServerConfig) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			sw := writer
			if serverConfig.ShouldSignData() {
				agentSign := request.Header.Get("HashSHA256")
				body, err := io.ReadAll(request.Body)
				if err != nil {
					http.Error(writer, fmt.Sprintf("failed read request body: %v", err), http.StatusInternalServerError)
					return
				}
				sign := createSign(body, serverConfig.SecretKey)
				encodeSign := base64.StdEncoding.EncodeToString(sign)
				if len(body) > 0 && agentSign != encodeSign {
					http.Error(writer, "invalid sign", http.StatusBadRequest)
					return
				}

				sw = newSignWriter(writer, serverConfig.SecretKey)
				if err := request.Body.Close(); err != nil {
					logger.Log.Warnf("failed close body in signature middleware: %v", err)
				}
				request.Body = io.NopCloser(bytes.NewReader(body))
			}

			handler.ServeHTTP(sw, request)
		})
	}
}

func createSign(body []byte, secretKey string) []byte {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(body)
	return h.Sum(nil)
}

func newSignWriter(w http.ResponseWriter, secretKey string) *SignWriter {
	return &SignWriter{
		ResponseWriter: w,
		secretKey:      secretKey,
	}
}

type SignWriter struct {
	http.ResponseWriter
	secretKey string
}

func (sw *SignWriter) Write(data []byte) (int, error) {
	sign := createSign(data, sw.secretKey)
	encodeSign := base64.StdEncoding.EncodeToString(sign)
	sw.ResponseWriter.Header().Set("HashSHA256", encodeSign)

	written, err := sw.ResponseWriter.Write(data)
	if err != nil {
		return written, fmt.Errorf("failed write sign data: %w", err)
	}

	return written, nil
}
