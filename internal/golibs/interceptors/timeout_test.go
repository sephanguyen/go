package interceptors

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const timeout = time.Millisecond * 500

type timeoutPingService struct {
	pb_testproto.TestServiceServer
}

// Ping expects to finish work in time.
func (s *timeoutPingService) Ping(ctx context.Context, fastPing *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	shortWorkTime := time.Millisecond * time.Duration(fastPing.SleepTimeMs)
	select {
	case <-time.After(shortWorkTime):
		return &pb_testproto.PingResponse{}, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// PingError expects to time out before finishing work.
func (s *timeoutPingService) PingError(ctx context.Context, slowPing *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	longWorkTime := time.Millisecond * time.Duration(slowPing.SleepTimeMs)
	select {
	case <-time.After(longWorkTime):
		return nil, fmt.Errorf("expected timeout after %v, but got %v", timeout, longWorkTime)
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

func TestTimeoutUnaryServerInterceptor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	addr := "127.0.0.1:42131"
	lis, err := net.Listen("tcp", addr)

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(TimeoutUnaryServerInterceptor(timeout)))
	defer s.GracefulStop()

	ts := &timeoutPingService{
		TestServiceServer: &grpc_testing.TestPingService{T: t},
	}
	pb_testproto.RegisterTestServiceServer(s, ts)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("s.Serve: %v", err)
		}
	}()

	clientConn, err := grpc.Dial(addr, grpc.WithBlock(), grpc.WithInsecure())
	require.NoError(t, err)
	cli := pb_testproto.NewTestServiceClient(clientConn)

	_, err = cli.Ping(ctx, &pb_testproto.PingRequest{SleepTimeMs: 100})
	assert.NoError(t, err)

	_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 1000})
	assert.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
}

func TestTimeoutUnaryServerInterceptorV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Setup server
	addr := "127.0.0.1:42132"
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	cfg := configs.GRPCHandlerTimeoutV2(map[string]time.Duration{
		"default": time.Millisecond * 500,
		"/mwitkow.testproto.TestService/PingError": time.Millisecond * 100,
	})
	it := TimeoutUnaryServerInterceptorV2(zap.NewNop(), cfg)
	require.NoError(t, err)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(it))
	defer s.GracefulStop()
	ts := &timeoutPingService{
		TestServiceServer: &grpc_testing.TestPingService{T: t},
	}
	pb_testproto.RegisterTestServiceServer(s, ts)
	go func() {
		serveErr := s.Serve(lis)
		require.NoError(t, serveErr)
	}()

	// Setup client and run test
	clientConn, err := grpc.Dial(addr,
		grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(time.Second*5),
	)
	require.NoError(t, err)
	cli := pb_testproto.NewTestServiceClient(clientConn)

	t.Run("Ping <500ms should not timeout", func(t *testing.T) {
		_, err = cli.Ping(ctx, &pb_testproto.PingRequest{SleepTimeMs: 100})
		require.NoError(t, err)
	})

	t.Run("Ping >500ms should timeout", func(t *testing.T) {
		_, err = cli.Ping(ctx, &pb_testproto.PingRequest{SleepTimeMs: 600})
		require.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
	})

	t.Run("PingError <100ms should not timeout", func(t *testing.T) {
		_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 50})
		// note that because PingError always return an error
		// but that doesnt mean that it's not working as expected
		require.EqualError(t, err, "rpc error: code = Unknown desc = expected timeout after 500ms, but got 50ms")
	})

	t.Run("PingError >100ms should timeout", func(t *testing.T) {
		_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 150})
		require.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
	})
}

func TestTimeoutUnaryServerInterceptorV2_NoDefaultTimeout(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Setup server
	addr := "127.0.0.1:42133"
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	cfg := configs.GRPCHandlerTimeoutV2(map[string]time.Duration{
		"default": -1,
		"/mwitkow.testproto.TestService/PingError": time.Millisecond * 100,
	})
	it := TimeoutUnaryServerInterceptorV2(zap.NewNop(), cfg)
	require.NoError(t, err)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(it))
	defer s.GracefulStop()
	ts := &timeoutPingService{
		TestServiceServer: &grpc_testing.TestPingService{T: t},
	}
	pb_testproto.RegisterTestServiceServer(s, ts)
	go func() {
		serveErr := s.Serve(lis)
		require.NoError(t, serveErr)
	}()

	// Setup client and run test
	clientConn, err := grpc.Dial(addr,
		grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(time.Second*5),
	)
	require.NoError(t, err)
	cli := pb_testproto.NewTestServiceClient(clientConn)

	t.Run("Ping >1s should not timeout", func(t *testing.T) {
		_, err = cli.Ping(ctx, &pb_testproto.PingRequest{SleepTimeMs: 1000})
		require.NoError(t, err)
	})

	t.Run("PingError <100ms should not timeout", func(t *testing.T) {
		_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 50})
		// note that because PingError always return an error
		// but that doesnt mean that it's not working as expected
		require.EqualError(t, err, "rpc error: code = Unknown desc = expected timeout after 500ms, but got 50ms")
	})
	t.Run("PingError >100ms should timeout", func(t *testing.T) {
		_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 150})
		require.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
	})
}

func TestTimeoutUnaryServerInterceptorV2_NoTimeoutSpecificAPI(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Setup server
	addr := "127.0.0.1:42134"
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	cfg := configs.GRPCHandlerTimeoutV2(map[string]time.Duration{
		"default": time.Millisecond * 500,
		"/mwitkow.testproto.TestService/PingError": -1,
	})
	it := TimeoutUnaryServerInterceptorV2(zap.NewNop(), cfg)
	require.NoError(t, err)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(it))
	defer s.GracefulStop()
	ts := &timeoutPingService{
		TestServiceServer: &grpc_testing.TestPingService{T: t},
	}
	pb_testproto.RegisterTestServiceServer(s, ts)
	go func() {
		serveErr := s.Serve(lis)
		require.NoError(t, serveErr)
	}()

	// Setup client and run test
	clientConn, err := grpc.Dial(addr,
		grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(time.Second*5),
	)
	require.NoError(t, err)
	cli := pb_testproto.NewTestServiceClient(clientConn)

	t.Run("Ping >500ms should timeout", func(t *testing.T) {
		_, err = cli.Ping(ctx, &pb_testproto.PingRequest{SleepTimeMs: 600})
		require.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
	})
	t.Run("PingError should not timeout", func(t *testing.T) {
		_, err = cli.PingError(ctx, &pb_testproto.PingRequest{SleepTimeMs: 1000})
		require.EqualError(t, err, "rpc error: code = Unknown desc = expected timeout after 500ms, but got 1s")
	})
}
