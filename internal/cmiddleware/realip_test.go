package cmiddleware

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRealIP(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		client := resty.New()
		request := client.NewRequest()

		middleware := WithRealIP()
		middleware(client, request)

		expectedIP, err := externalIP()
		require.NoError(t, err)
		got := request.Header.Get("X-Real-IP")
		assert.Equal(t, expectedIP, got)
	})
}
