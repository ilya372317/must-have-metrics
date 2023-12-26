package middleware

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_isCorrectSigned(t *testing.T) {
	type args struct {
		secretKey   string
		body        []byte
		bodyForSign []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success empty key case",
			args: args{
				secretKey:   "",
				body:        []byte(""),
				bodyForSign: []byte(""),
			},
			want: true,
		},
		{
			name: "success correct sign case",
			args: args{
				secretKey:   "1234",
				body:        []byte("Ilya"),
				bodyForSign: []byte("Ilya"),
			},
			want: true,
		},
		{
			name: "invalid sign case",
			args: args{
				secretKey:   "1234",
				body:        []byte("Ilya"),
				bodyForSign: []byte("Otinov"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(tt.args.body))
			sign := signature.CreateSign(tt.args.bodyForSign, tt.args.secretKey)
			req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(sign))

			cnfg := &config.ServerConfig{
				SecretKey: tt.args.secretKey,
			}
			got, err := isCorrectSigned(cnfg, req)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
