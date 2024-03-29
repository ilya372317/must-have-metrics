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

	"github.com/go-resty/resty/v2"
)

const (
	hashLengthTimes     = 2
	extraBytesForCipher = 2
)

// WithRSACrypt Encrypt request body by public key.
// Public key retrieve from file in publicKeyPath argument
func WithRSACrypt(publicKeyPath string) resty.RequestMiddleware {
	return func(client *resty.Client, request *resty.Request) error {
		body, ok := request.Body.([]byte)
		if !ok {
			return fmt.Errorf("request body expected to be byte slice")
		}
		publicKeyData, err := getPublicKeyData(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed get public key data: %w", err)
		}

		block, _ := pem.Decode(publicKeyData)
		publickKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed parse given public key: %w", err)
		}

		cipherBlockSize := (publickKey.Size()) - (sha256.Size * hashLengthTimes) - extraBytesForCipher
		cipherData := make([]byte, 0)
		message := body

		for len(message) > 0 {
			var chunk []byte
			if len(message) > cipherBlockSize {
				chunk = message[:cipherBlockSize]
				message = message[cipherBlockSize:]
			} else {
				chunk = message
				message = nil
			}

			encryptData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publickKey, chunk, []byte(""))
			if err != nil {
				return fmt.Errorf("failed chipher request body: %w", err)
			}

			cipherData = append(cipherData, encryptData...)
		}

		request.SetBody(base64.StdEncoding.EncodeToString(cipherData))
		return nil
	}
}

func getPublicKeyData(publicKeyPath string) ([]byte, error) {
	publicKeyContent, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return []byte(""), fmt.Errorf("failed get public key content: %w", err)
	}

	return publicKeyContent, nil
}
