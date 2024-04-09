package agent

import (
	"context"
	"fmt"

	"github.com/ilya372317/must-have-metrics/pgk/externalip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
			return fmt.Errorf("failed resolve ip address: %w", err)
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
