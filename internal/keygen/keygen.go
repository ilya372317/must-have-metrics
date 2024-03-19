package keygen

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func GenerateRSAKeys(keysDir string, keySize int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("failed generate key: %w", err)
	}

	publicKey := privateKey.PublicKey

	var privateKeyPEM bytes.Buffer
	var publicKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return fmt.Errorf("failed encode private key: %w", err)
	}

	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
	})
	if err != nil {
		return fmt.Errorf("failed encode public key: %w", err)
	}

	if _, err = os.Stat(keysDir); os.IsNotExist(err) {
		err = os.Mkdir(keysDir, 0750)
		if err != nil {
			return fmt.Errorf("failed create keys dir: %w", err)
		}
	}

	privateFile, err := os.OpenFile(keysDir+"/private-key.pem", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0650)
	if err != nil {
		return fmt.Errorf("failed open private file: %w", err)
	}
	defer func() {
		_ = privateFile.Close()
	}()
	publicFile, err := os.OpenFile(keysDir+"/public-key.pem", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0650)
	if err != nil {
		return fmt.Errorf("failed open public file: %w", err)
	}
	defer func() {
		_ = publicFile.Close()
	}()

	_, err = publicFile.Write(publicKeyPEM.Bytes())
	if err != nil {
		return fmt.Errorf("failed write data to public file: %w", err)
	}

	_, err = privateFile.Write(privateKeyPEM.Bytes())
	if err != nil {
		return fmt.Errorf("failed write data to private file: %w", err)
	}

	return nil
}
