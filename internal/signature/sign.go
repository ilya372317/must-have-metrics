package signature

import (
	"crypto/hmac"
	"crypto/sha256"
)

// CreateSign from given body and secret key create signature.
func CreateSign(body []byte, secretKey string) []byte {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(body)
	return h.Sum(nil)
}
