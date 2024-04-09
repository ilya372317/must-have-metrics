package cmiddleware

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/pgk/externalip"
)

func WithRealIP() resty.RequestMiddleware {
	return func(c *resty.Client, r *resty.Request) error {
		ipStr, err := externalip.Get()
		if err != nil {
			return fmt.Errorf("failed resolve client ip address: %w", err)
		}
		r.Header.Set("X-Real-IP", ipStr)

		return nil
	}
}
