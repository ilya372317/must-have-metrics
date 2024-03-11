package cmiddleware

import (
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/signature"
)

// WithSignature middleware for attaching request signature to headers.
func WithSignature(secretKey string) resty.RequestMiddleware {
	return func(client *resty.Client, request *resty.Request) error {
		body, ok := request.Body.(string)
		if !ok {
			return fmt.Errorf("request body not a string")
		}
		sign := signature.CreateSign([]byte(body), secretKey)
		encodeSing := base64.StdEncoding.EncodeToString(sign)
		request.Header.Set("HashSHA256", encodeSing)
		return nil
	}
}
