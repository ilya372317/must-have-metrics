package cmiddleware

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/pgk/externalip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRealIP(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		client := resty.New()
		request := client.NewRequest()

		middleware := WithRealIP()
		err := middleware(client, request)
		require.NoError(t, err)

		expectedIP, err := externalip.Get()
		require.NoError(t, err)
		got := request.Header.Get("X-Real-IP")
		assert.Equal(t, expectedIP, got)
	})
}
