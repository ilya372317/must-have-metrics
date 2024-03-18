package cmiddleware

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
)

var bodyIsNotStringErr = fmt.Errorf("request body not a string")

// WithCompress middleware for compressing requests sent to the server.
func WithCompress() resty.RequestMiddleware {
	return func(client *resty.Client, request *resty.Request) error {
		body, ok := request.Body.(string)
		if !ok {
			return bodyIsNotStringErr
		}

		compressedBody, err := compress.Do([]byte(body))
		if err != nil {
			return fmt.Errorf("failed compress body: %w", err)
		}
		request.SetBody(compressedBody)
		request.SetHeader("Content-Encoding", "gzip")

		return nil
	}
}
