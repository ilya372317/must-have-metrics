package cmiddleware

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithCompress(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		contentEncoding string
		wantErr         bool
	}{
		{
			name:            "simple success case",
			body:            "test body",
			contentEncoding: "gzip",
			wantErr:         false,
		},
		{
			name:            "empty body case",
			body:            "",
			contentEncoding: "gzip",
			wantErr:         false,
		},
		{
			name:            "not string body given",
			body:            "not-string",
			contentEncoding: "none",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := WithCompress()
			client := resty.New()
			request := client.NewRequest()
			request.SetHeader("Content-Encoding", "none")
			if tt.body == "not-string" {
				request.SetBody([]byte(tt.body))
			} else {
				request.SetBody(tt.body)
			}
			err := middleware(client, request)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			compressedBody, err := compress.Do([]byte(tt.body))
			require.NoError(t, err)
			assert.Equal(t, compressedBody, request.Body.([]byte))
			assert.Equal(t, tt.contentEncoding, request.Header.Get("Content-Encoding"))
		})
	}
}
