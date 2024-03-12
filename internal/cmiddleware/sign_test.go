package cmiddleware

import (
	"encoding/base64"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSignature(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		body      string
	}{
		{
			name:      "simple success case",
			secretKey: "test123",
			body:      "testBody",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := WithSignature(tt.secretKey)
			sign := signature.CreateSign([]byte(tt.body), tt.secretKey)
			encodeSing := base64.StdEncoding.EncodeToString(sign)
			client := resty.New()
			request := client.NewRequest()
			request.SetBody(tt.body)
			err := middleware(client, request)
			require.NoError(t, err)

			assert.Equal(t, encodeSing, request.Header.Get("HashSHA256"))
		})
	}
}
