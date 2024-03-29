package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTrustedSubnet(t *testing.T) {
	tests := []struct {
		name          string
		clientIP      string
		trustedSubnet string
		wantCode      int
	}{
		{
			name:          "invalid ip client given",
			clientIP:      "not-ip-address",
			trustedSubnet: "127.0.0.0/24",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "invalid server subnet config",
			clientIP:      "127.0.0.1",
			trustedSubnet: "not-a-subnet-string",
			wantCode:      http.StatusInternalServerError,
		},
		{
			name:          "success C class subnet",
			clientIP:      "127.0.0.1",
			trustedSubnet: "127.0.0.0/24",
			wantCode:      http.StatusOK,
		},
		{
			name:          "success B class subnet",
			clientIP:      "127.0.200.192",
			trustedSubnet: "127.0.0.0/16",
			wantCode:      http.StatusOK,
		},
		{
			name:          "success A class subnet",
			clientIP:      "127.180.221.95",
			trustedSubnet: "127.0.0.0/8",
			wantCode:      http.StatusOK,
		},
		{
			name:          "not valid ip client for C class subnet",
			clientIP:      "127.0.1.1",
			trustedSubnet: "127.0.0.0/24",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "not valid ip client for B class subnet",
			clientIP:      "127.1.1.1",
			trustedSubnet: "127.0.0.0/16",
			wantCode:      http.StatusForbidden,
		},
		{
			name:          "not valid ip client for A class subnet",
			clientIP:      "128.1.1.1",
			trustedSubnet: "127.0.0.1/8",
			wantCode:      http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.Header.Set("X-Real-IP", tt.clientIP)
			w := httptest.NewRecorder()
			cnfg := config.ServerConfig{
				TrustedSubnet: tt.trustedSubnet,
			}

			middleware := WithTrustedSubnet(&cnfg)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			handler.ServeHTTP(w, r)

			res := w.Result()

			err := res.Body.Close()
			require.NoError(t, err)

			got := res.StatusCode

			assert.Equal(t, tt.wantCode, got)
		})
	}
}
