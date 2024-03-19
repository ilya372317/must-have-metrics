package cmiddleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/keygen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keysDir = "/tmp/cipher-test"
const publicKeyPath = "public-key.pem"

func TestMain(m *testing.M) {
	const keySize = 4096
	if _, err := os.Stat(keysDir); os.IsNotExist(err) {
		err = os.Mkdir(keysDir, 0750)
		if err != nil {
			panic("failed create folder for test cipher")
		}
	}
	if err := keygen.GenerateRSAKeys(keysDir, keySize); err != nil {
		panic(fmt.Errorf("failed generate rsa keys: %w", err))
	}

	m.Run()

	if err := os.RemoveAll(keysDir); err != nil {
		panic(fmt.Errorf("failed clean up rsa keys in path %s. do it manualy: %w", keysDir, err))
	}
}

func TestWithRSACipher(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    bool
		body       string
		pathToKeys string
	}{
		{
			name:       "simple success case",
			wantErr:    false,
			body:       "test body 123",
			pathToKeys: keysDir + "/" + publicKeyPath,
		},
		{
			name:       "given invalid path to public key",
			wantErr:    true,
			body:       "test 123",
			pathToKeys: "",
		},
		{
			name:    "very long body",
			wantErr: false,
			body: "test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks" +
				" test 123 it is very long string and it should be split for chunks",
			pathToKeys: keysDir + "/" + publicKeyPath,
		},
	}

	hasher := sha256.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			request := client.NewRequest()
			request.SetBody(tt.body)
			middleware := WithRSACipher(tt.pathToKeys)
			err := middleware(client, request)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			privateKeyData, err := os.ReadFile(keysDir + "/private-key.pem")
			require.NoError(t, err)
			block, _ := pem.Decode(privateKeyData)
			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			require.NoError(t, err)
			bodyData, err := base64.StdEncoding.DecodeString(request.Body.(string))
			require.NoError(t, err)
			blockSize := privateKey.Size()
			var decryptedData []byte
			for len(bodyData) > 0 {
				var chunk []byte
				if len(bodyData) > blockSize {
					chunk = bodyData[:blockSize]
					bodyData = bodyData[blockSize:]
				} else {
					chunk = bodyData
					bodyData = nil
				}
				uncipherData, err := rsa.DecryptOAEP(hasher, rand.Reader, privateKey, chunk, []byte(""))
				require.NoError(t, err)
				decryptedData = append(decryptedData, uncipherData...)
			}

			assert.Equal(t, tt.body, string(decryptedData))
		})
	}
}
