package server

import (
	"context"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/pgk/externalip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestWithTrustedSubnet(t *testing.T) {
	validIP, err := externalip.Get()
	mySubnet := validIP + "/24"
	require.NoError(t, err)
	type requestData struct {
		subnet      string
		realIP      string
		hasMetadata bool
	}
	tests := []struct {
		name     string
		request  requestData
		wantCode codes.Code
		wantErr  bool
	}{
		{
			request: requestData{
				hasMetadata: false,
				subnet:      mySubnet,
			},
			name:     "metadata not given",
			wantCode: codes.Aborted,
			wantErr:  true,
		},
		{
			request: requestData{
				subnet:      "",
				hasMetadata: false,
			},
			name:     "success case with not trusted subnet check",
			wantCode: 0,
			wantErr:  false,
		},
		{
			request: requestData{
				subnet:      mySubnet,
				hasMetadata: false,
				realIP:      "",
			},
			name:     "empty real ip case",
			wantCode: codes.Aborted,
			wantErr:  true,
		},
		{
			request: requestData{
				subnet:      "invalid-subnet",
				hasMetadata: true,
				realIP:      validIP,
			},
			name:     "invalid subnet given",
			wantCode: codes.Internal,
			wantErr:  true,
		},
		{
			request: requestData{
				subnet:      mySubnet,
				hasMetadata: true,
				realIP:      "invalid-real-ip",
			},
			name:     "invalid real ip given from client",
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			request: requestData{
				subnet:      mySubnet,
				hasMetadata: true,
				realIP:      validIP,
			},
			name:    "trusted subnet success case",
			wantErr: false,
		},
		{
			request: requestData{
				subnet:      "192.168.0.1/24",
				hasMetadata: true,
				realIP:      "192.168.1.1",
			},
			name:     "client ip not in valid subnet",
			wantCode: codes.Aborted,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.New(make(map[string]string))
			md.Set("X-Real-IP", tt.request.realIP)
			var ctx context.Context
			if tt.request.hasMetadata {
				ctx = metadata.NewIncomingContext(context.Background(), md)
			} else {
				ctx = context.Background()
			}
			handler := func(ctx context.Context, req any) (any, error) {
				return "success", nil
			}
			interceptor := WithTrustedSubnet(&config.ServerConfig{TrustedSubnet: tt.request.subnet})

			res, err := interceptor(
				ctx,
				"{}",
				&grpc.UnaryServerInfo{},
				handler,
			)

			if tt.wantErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, e.Code())
				return
			} else {
				require.NoError(t, err)
			}

			strResp, ok := res.(string)
			assert.True(t, ok)
			assert.Equal(t, "success", strResp)
		})
	}
}
