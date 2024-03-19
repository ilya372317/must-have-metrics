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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/keygen"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keysDir = "/tmp/cipher-test"
const publicKeyPath = "public-key.pem"
const privateKeyPath = "private-key.pem"

func TestMain(m *testing.M) {
	if err := keygen.GenerateRSAKeys(keysDir, 4096); err != nil {
		panic(fmt.Errorf("failed generate rsa keys for test: %w", err))
	}
	if err := logger.Init(); err != nil {
		panic(fmt.Errorf("failed init logger for testing: %w", err))
	}

	m.Run()
	if err := os.RemoveAll(keysDir); err != nil {
		panic(fmt.Errorf("failed clean up rsa keys in path %s. do it manualy: %w", keysDir, err))
	}
}

func TestWithRSADecrypt(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantCode   int
		encodeBody bool
	}{
		{
			name:       "simple success case",
			body:       "test 123",
			wantCode:   http.StatusOK,
			encodeBody: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publicKeyData, err := os.ReadFile(keysDir + "/" + publicKeyPath)
			require.NoError(t, err)
			block, _ := pem.Decode(publicKeyData)
			publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
			require.NoError(t, err)
			blockSize := publicKey.Size() - sha256.Size*2 - 2
			byteBody := []byte(tt.body)
			var cryptedBody []byte
			for len(byteBody) > 0 {
				var chunk []byte
				if len(byteBody) > blockSize {
					chunk = byteBody[:blockSize]
					byteBody = byteBody[blockSize:]
				} else {
					chunk = byteBody
					byteBody = nil
				}
				cryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, chunk, []byte(""))
				require.NoError(t, err)
				cryptedBody = append(cryptedBody, cryptedChunk...)
			}
			if tt.encodeBody {
				cryptedBody = []byte(base64.StdEncoding.EncodeToString(cryptedBody))
			}

			r := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(cryptedBody))
			w := httptest.NewRecorder()

			middleware := WithRSADecrypt(keysDir + "/" + privateKeyPath)
			handler := middleware(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				rBody, err := io.ReadAll(request.Body)
				require.NoError(t, err)
				_, err = responseWriter.Write(rBody)
				require.NoError(t, err)
				return
			}))
			handler.ServeHTTP(w, r)

			res := w.Result()
			defer func() {
				err = res.Body.Close()
				require.NoError(t, err)
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode)
			assert.Equal(t, tt.body, string(resBody))
		})
	}
}
