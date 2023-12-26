package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_createSign(t *testing.T) {
	type argument struct {
		src,
		key string
	}
	tests := []struct {
		name string
		arg  argument
	}{
		{
			name: "simple success case",
			arg: argument{
				src: "Ilya Otinov",
				key: "1234567",
			},
		},
		{
			name: "long success case",
			arg: argument{
				src: "It is very long and complex string, and i expect hash to be equal anyway",
				key: "my-name-is-ilya",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateSign([]byte(tt.arg.src), tt.arg.key)

			h := hmac.New(sha256.New, []byte(tt.arg.key))
			h.Write([]byte(tt.arg.src))
			assert.Equal(t, h.Sum(nil), got)
		})
	}
}

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
			sign := CreateSign(tt.args.bodyForSign, tt.args.secretKey)
			req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(sign))

			cnfg := &config.ServerConfig{
				SecretKey: tt.args.secretKey,
			}
			got, err := IsCorrectSigned(cnfg, req)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
