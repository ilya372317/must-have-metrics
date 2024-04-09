package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestWithRealIP(t *testing.T) {
	interceptor := WithRealIP()
	invoker := func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.True(t, len(md.Get("X-Real-IP")) > 0)
		assert.Regexp(t, "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$", md.Get("X-Real-IP")[0])

		return nil
	}
	ctx := context.Background()
	err := interceptor(
		ctx,
		"method",
		"req",
		"reply",
		&grpc.ClientConn{},
		invoker,
	)
	require.NoError(t, err)
}
