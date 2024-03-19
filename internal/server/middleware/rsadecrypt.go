package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ilya372317/must-have-metrics/internal/logger"
)

const (
	hashLengthTimes     = 2
	extraBytesForCipher = 2
)

// WithRSADecrypt middleware for decrypt given request body be RSA algo.
func WithRSADecrypt(privateKeyPath string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			privateKeyData, err := os.ReadFile(privateKeyPath)
			if err != nil {
				http.Error(
					w,
					fmt.Errorf("failed get key for decryption: %w", err).Error(),
					http.StatusInternalServerError,
				)
				return
			}
			block, _ := pem.Decode(privateKeyData)
			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				err = fmt.Errorf("invalid private key content: %w", err)
				logger.Log.Error(err.Error())
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			requestData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(
					w,
					fmt.Errorf("failed read request body: %w", err).Error(),
					http.StatusInternalServerError,
				)
				return
			}
			base64RequestData, err := base64.StdEncoding.DecodeString(string(requestData))
			if err != nil {
				http.Error(
					w,
					fmt.Errorf("crypted body expected be in base64 encoding").Error(),
					http.StatusBadRequest,
				)
				return
			}
			blockSize := privateKey.Size()
			var decryptedData []byte
			for len(base64RequestData) > 0 {
				var chunk []byte
				if len(base64RequestData) > blockSize {
					chunk = base64RequestData[:blockSize]
					base64RequestData = base64RequestData[blockSize:]
				} else {
					chunk = base64RequestData
					base64RequestData = nil
				}
				uncryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, chunk, []byte(""))
				if err != nil {
					err = fmt.Errorf("failed decrypt request body: %w", err)
					logger.Log.Error(err.Error())
					http.Error(
						w,
						err.Error(),
						http.StatusInternalServerError,
					)
					return
				}
				decryptedData = append(decryptedData, uncryptedData...)
			}
			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			h.ServeHTTP(w, r)
		})
	}
}
