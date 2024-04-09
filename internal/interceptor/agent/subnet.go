package agent

import (
	"context"

	"github.com/ilya372317/must-have-metrics/pgk/externalip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func WithRealIP() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ipStr, err := externalip.Get()
		if err != nil {
			return status.Errorf(codes.Internal, "failed resolve ip address: %v", err)
		}

		var md metadata.MD
		var ok bool
		md, ok = metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(map[string]string{})
		}

		md.Set("X-Real-IP", ipStr)

		newCtx := metadata.NewOutgoingContext(ctx, md)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}
