package server

import (
	"context"
	"net"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func WithTrustedSubnet(serverConfig *config.ServerConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		if !serverConfig.ShouldCheckIP() {
			return handler(ctx, req)
		}
		var realIP string
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			realIPs := md.Get("X-Real-IP")
			if len(realIPs) > 0 {
				realIP = realIPs[0]
			}
		}

		if len(realIP) == 0 {
			return nil, status.Error(codes.Aborted, "invalid X-Real-IP from grpc client given")
		}

		_, trustedIPNet, err := net.ParseCIDR(serverConfig.TrustedSubnet)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid server trusted ip configuration: %v", err)
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid real client ip given")
		}

		if !trustedIPNet.Contains(ip) {
			return nil, status.Errorf(codes.Aborted, "client ip not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
